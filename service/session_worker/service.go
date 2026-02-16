package session_worker

import (
	"context"
	"time"

	"msp-admin-service/entity"

	"github.com/pkg/errors"
	"github.com/txix-open/bgjob"
	"github.com/txix-open/isp-kit/bgjobx/handler"
	"github.com/txix-open/isp-kit/log"
)

const (
	defaultRetryTimeout = 5 * time.Minute
	rescheduleInterval  = 1 * time.Minute
)

type TokenTransactionRunner interface {
	TokenTransaction(ctx context.Context, tx func(ctx context.Context, tx TokenTransaction) error) error
}

type TokenTransaction interface {
	All(ctx context.Context) ([]entity.Token, error)
	SetExpiredStatusByIds(ctx context.Context, ids []int) error
}

type Service struct {
	logger   log.Logger
	txRunner TokenTransactionRunner
}

func NewExpireSessionWorker(
	logger log.Logger,
	txRunner TokenTransactionRunner,
) Service {
	return Service{
		logger:   logger,
		txRunner: txRunner,
	}
}

func (w Service) Handle(ctx context.Context, _ bgjob.Job) handler.Result {
	ctx = log.ToContext(ctx, log.String("worker", "expireSessionWorker"))
	w.logger.Debug(ctx, "begin work")

	err := w.do(ctx)
	if err != nil {
		return handler.Retry(defaultRetryTimeout, errors.WithMessage(err, "expireSessionWorker do"))
	}

	w.logger.Debug(ctx, "end work")
	return handler.Reschedule(handler.ByAfterTime(rescheduleInterval, time.Now()))
}

func (w Service) do(ctx context.Context) error {
	err := w.txRunner.TokenTransaction(ctx, func(ctx context.Context, tx TokenTransaction) error {
		tokens, err := tx.All(ctx)
		if err != nil {
			return errors.WithMessage(err, "get all tokens")
		}

		expiredTokenIds := make([]int, 0)
		now := time.Now().UTC()
		for _, token := range tokens {
			if now.After(token.ExpiredAt) {
				expiredTokenIds = append(expiredTokenIds, token.Id)
			}
		}

		if len(expiredTokenIds) == 0 {
			return nil
		}

		err = tx.SetExpiredStatusByIds(ctx, expiredTokenIds)
		if err != nil {
			return errors.WithMessage(err, "set expired status by id")
		}

		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "token transaction")
	}

	return nil
}
