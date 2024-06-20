package inactive_worker

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/bgjob"
	"github.com/txix-open/isp-kit/log"
	"msp-admin-service/conf"
	"msp-admin-service/entity"
)

const (
	defaultRetryTimeout = 5 * time.Minute
)

type AuditRepo interface {
	SaveAuditAsync(ctx context.Context, userId int64, message string, event string)
}

type TokenRepo interface {
	LastAccessNotBlockedUsers(ctx context.Context) (map[int64]time.Time, error)
}

type UserRepo interface {
	Block(ctx context.Context, userId int) (*entity.User, error)
}

type LdapService interface {
	SyncGroups(ctx context.Context, user entity.User) error
}

type UserRoleRepo interface {
	UpsertUserRoleLinks(ctx context.Context, id int, roleIds []int) error
}

type Service struct {
	tokenRepo    TokenRepo
	userRepo     UserRepo
	auditRepo    AuditRepo
	userRoleRepo UserRoleRepo
	ldapService  LdapService
	threshold    time.Duration
	syncPeriod   time.Duration
	logger       log.Logger
}

func NewInactiveBlocker(
	tokensRepo TokenRepo,
	userRepo UserRepo,
	auditRepo AuditRepo,
	userRoleRepo UserRoleRepo,
	ldapService LdapService,
	config conf.BlockInactiveWorker,
	logger log.Logger,
) Service {
	return Service{
		tokenRepo:    tokensRepo,
		userRepo:     userRepo,
		auditRepo:    auditRepo,
		userRoleRepo: userRoleRepo,
		ldapService:  ldapService,
		threshold:    time.Duration(config.DaysThreshold) * 24 * time.Hour,
		syncPeriod:   time.Minute * time.Duration(config.RunIntervalInMinutes),
		logger:       logger,
	}
}

func (w Service) Handle(ctx context.Context, _ bgjob.Job) bgjob.Result {
	ctx = log.ToContext(ctx, log.String("worker", "inactiveBlocker"))
	w.logger.Debug(ctx, "begin work")
	err := w.do(ctx)
	if err != nil {
		return bgjob.Retry(defaultRetryTimeout, errors.WithMessage(err, "sync state with worker"))
	}
	w.logger.Debug(ctx, "end work")
	return bgjob.Reschedule(w.syncPeriod)
}

func (w Service) do(ctx context.Context) error {
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
		user, err := w.userRepo.Block(ctx, int(userId))
		if err != nil {
			return errors.WithMessagef(err, "block user %d", userId)
		}

		err = w.userRoleRepo.UpsertUserRoleLinks(ctx, int(userId), []int{})
		if err != nil {
			return errors.WithMessage(err, "empty user roles links")
		}

		err = w.ldapService.SyncGroups(ctx, *user)
		if err != nil {
			w.logger.Error(ctx, errors.WithMessage(err, "remove ldap groups"))
		}

		w.auditRepo.SaveAuditAsync(ctx, userId, "Блокировка неактивной УЗ", entity.EventUserChanged)
	}

	return nil
}
