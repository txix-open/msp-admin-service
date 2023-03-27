package worker

import (
	"context"
	"time"

	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
)

type AuditRepo interface {
	SaveAuditAsync(ctx context.Context, userId int64, message string)
}

type TokenRepo interface {
	LastAccessNotBlockedUsers(ctx context.Context) (map[int64]time.Time, error)
}

type UserRepo interface {
	Block(ctx context.Context, userId int) error
}

type InactiveBlocker struct {
	tokenRepo TokenRepo
	userRepo  UserRepo
	auditRepo AuditRepo
	threshold time.Duration
	logger    log.Logger
}

func NewInactiveBlocker(
	tokensRepo TokenRepo,
	userRepo UserRepo,
	auditRepo AuditRepo,
	inactiveThresholdInDays int,
	logger log.Logger,
) InactiveBlocker {
	return InactiveBlocker{
		tokenRepo: tokensRepo,
		userRepo:  userRepo,
		auditRepo: auditRepo,
		threshold: time.Duration(inactiveThresholdInDays) * 24 * time.Hour,
		logger:    logger,
	}
}

func (w InactiveBlocker) Do(ctx context.Context) {
	ctx = log.ToContext(ctx, log.String("worker", "inactiveBlocker"))
	w.logger.Debug(ctx, "begin work")
	err := w.do(ctx)
	if err != nil {
		w.logger.Error(ctx, "error", log.Any("error", err))
	}
	w.logger.Debug(ctx, "end work")
}

func (w InactiveBlocker) do(ctx context.Context) error {
	lastAccess, err := w.tokenRepo.LastAccessNotBlockedUsers(ctx)
	if err != nil {
		return errors.WithMessage(err, "get last access times")
	}

	now := time.Now().UTC()
	for userId, lastAccess := range lastAccess {
		dur := now.Sub(lastAccess)
		if dur < w.threshold {
			continue
		}

		w.logger.Info(ctx, "block inactive user", log.Int64("userId", userId), log.Any("lastAccessTime", lastAccess))
		err := w.userRepo.Block(ctx, int(userId))
		if err != nil {
			return errors.WithMessagef(err, "block user %d", userId)
		}
		w.auditRepo.SaveAuditAsync(ctx, userId, "Блокировка неактивной УЗ")
	}

	return nil
}
