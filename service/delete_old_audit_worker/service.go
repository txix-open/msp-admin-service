package delete_old_audit_worker

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/bgjob"
	"github.com/txix-open/isp-kit/bgjobx/handler"
	"github.com/txix-open/isp-kit/log"
	"msp-admin-service/conf"
)

const (
	defaultRetryTimeout = 5 * time.Minute
)

type AuditRepository interface {
	DeleteUpToCreatedAt(ctx context.Context, createdAt time.Time) error
}

type Service struct {
	logger     log.Logger
	auditRep   AuditRepository
	syncPeriod time.Duration
	auditTTL   time.Duration
}

func NewService(
	logger log.Logger,
	auditRep AuditRepository,
	setting conf.AuditTTlSetting,
) Service {
	return Service{
		logger:     logger,
		auditRep:   auditRep,
		syncPeriod: time.Minute * time.Duration(setting.ExpireSyncPeriodInMin),
		auditTTL:   time.Minute * time.Duration(setting.TimeToLiveInMin),
	}
}

func (s Service) Handle(ctx context.Context, _ bgjob.Job) handler.Result {
	err := s.deleteOldAudit(ctx)
	if err != nil {
		return handler.Retry(defaultRetryTimeout, errors.WithMessage(err, "sync state with worker"))
	}

	return handler.Reschedule(s.syncPeriod)
}

func (s Service) deleteOldAudit(ctx context.Context) error {
	expirationDeadLine := time.Now().UTC().Add(-s.auditTTL)
	err := s.auditRep.DeleteUpToCreatedAt(ctx, expirationDeadLine)
	if err != nil {
		return errors.WithMessage(err, "delete audit")
	}

	return nil
}
