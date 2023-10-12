package service

import (
	"context"
	"time"

	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"msp-admin-service/conf"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
)

type AuditRepository interface {
	Insert(ctx context.Context, log entity.Audit) (int, error)
	All(ctx context.Context, limit int, offset int) ([]entity.Audit, error)
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
	go func() {
		isEnable, err := s.auditEventRep.IsEnable(context.Background(), event)
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

func (s Audit) All(ctx context.Context, limit int, offset int) (*domain.AuditResponse, error) {
	group, ctx := errgroup.WithContext(ctx)
	var tokens []entity.Audit
	var total int64
	var err error
	group.Go(func() error {
		tokens, err = s.auditRep.All(ctx, limit, offset)
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
	for _, token := range tokens {
		items = append(items, domain.Audit{
			Id:        token.Id,
			UserId:    token.UserId,
			Message:   token.Message,
			CreatedAt: token.CreatedAt,
		})
	}
	result := domain.AuditResponse{
		TotalCount: int(total),
		Items:      items,
	}

	return &result, nil
}
