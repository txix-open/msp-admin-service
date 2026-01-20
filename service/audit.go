package service

import (
	"context"
	"time"

	"msp-admin-service/conf"
	"msp-admin-service/domain"
	"msp-admin-service/entity"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
	"golang.org/x/sync/errgroup"
)

type AuditRepository interface {
	Insert(ctx context.Context, log entity.Audit) (int, error)
	All(ctx context.Context, req domain.AuditPageRequest) ([]entity.Audit, error)
	Count(ctx context.Context) (int64, error)
}

type AuditEventRepository interface {
	All(ctx context.Context) ([]entity.AuditEvent, error)
	Upsert(ctx context.Context, eventList []entity.AuditEvent) error
	IsEnable(ctx context.Context, event string) (bool, error)
}

type Audit struct {
	logger            log.Logger
	auditRep          AuditRepository
	auditEventRep     AuditEventRepository
	eventSetting      map[string]conf.AuditEventSetting
	expectedEventList map[string]bool
}

func NewAudit(
	ctx context.Context,
	logger log.Logger,
	auditRep AuditRepository,
	auditEventRep AuditEventRepository,
	settings []conf.AuditEventSetting,
) Audit {
	expectedEventList := map[string]bool{
		entity.EventSuccessLogin:  true,
		entity.EventErrorLogin:    true,
		entity.EventSuccessLogout: true,
		entity.EventRoleChanged:   true,
		entity.EventUserChanged:   true,
		entity.EventUserBlocked:   true,
	}

	eventName := make(map[string]conf.AuditEventSetting)
	for _, setting := range settings {
		expectedEvent := expectedEventList[setting.Event]
		if !expectedEvent {
			logger.Warn(ctx, "not expected audit event in remote config", log.String("event", setting.Event))
			continue
		}

		eventName[setting.Event] = setting
	}

	return Audit{
		logger:            logger,
		auditRep:          auditRep,
		auditEventRep:     auditEventRep,
		eventSetting:      eventName,
		expectedEventList: expectedEventList,
	}
}

func (s Audit) SaveAuditAsync(ctx context.Context, userId int64, message string, event string) {
	ctx = context.WithoutCancel(ctx)
	go func() {
		isEnable, err := s.auditEventRep.IsEnable(ctx, event)
		if err != nil {
			s.logger.Error(ctx, "check is enable audit event", log.Any("error", err))
			return
		}
		if !isEnable {
			return
		}

		audit := entity.Audit{
			UserId:    int(userId),
			Message:   message,
			Event:     event,
			CreatedAt: time.Now().UTC(),
		}
		_, err = s.auditRep.Insert(context.Background(), audit)
		if err != nil {
			s.logger.Error(ctx, "insert audit", log.Any("error", err))
		}
	}()
}

func (s Audit) All(ctx context.Context, req domain.AuditPageRequest) (*domain.AuditResponse, error) {
	group, ctx := errgroup.WithContext(ctx)
	var logs []entity.Audit
	var total int64
	var err error
	group.Go(func() error {
		logs, err = s.auditRep.All(ctx, req)
		if err != nil {
			return errors.WithMessage(err, "get all audit")
		}
		return nil
	})
	group.Go(func() error {
		total, err = s.auditRep.Count(ctx)
		if err != nil {
			return errors.WithMessage(err, "count all audit")
		}
		return nil
	})
	err = group.Wait()
	if err != nil {
		return nil, errors.WithMessage(err, "wait workers")
	}

	items := make([]domain.Audit, 0)
	for _, log := range logs {
		items = append(items, domain.Audit{
			Id:        log.Id,
			UserId:    log.UserId,
			Message:   log.Message,
			CreatedAt: log.CreatedAt,
		})
	}
	result := domain.AuditResponse{
		TotalCount: int(total),
		Items:      items,
	}

	return &result, nil
}

func (s Audit) Events(ctx context.Context) ([]domain.AuditEvent, error) {
	eventList, err := s.auditEventRep.All(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get all audit_event")
	}

	result := make([]domain.AuditEvent, len(eventList))
	for i, event := range eventList {
		result[i] = domain.AuditEvent{
			Event:   event.Event,
			Name:    s.eventSetting[event.Event].Name,
			Enabled: event.Enable,
		}
	}

	return result, nil
}

func (s Audit) SetEvents(ctx context.Context, req []domain.SetAuditEvent) error {
	if len(req) == 0 {
		return nil
	}

	eventList := make([]entity.AuditEvent, len(req))
	for i, event := range req {
		_, found := s.expectedEventList[event.Event]
		if !found {
			return domain.UnknownAuditEventError{Event: event.Event}
		}

		eventList[i] = entity.AuditEvent{
			Event:  event.Event,
			Enable: event.Enabled,
		}
	}

	err := s.auditEventRep.Upsert(ctx, eventList)
	if err != nil {
		return errors.WithMessagef(err, "upsert audit_event list")
	}

	return nil
}
