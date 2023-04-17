package service

import (
	"context"
	"time"

	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
)

type AuditRepo interface {
	Insert(ctx context.Context, log entity.Audit) (int, error)
	All(ctx context.Context, limit int, offset int) ([]entity.Audit, error)
	Count(ctx context.Context) (int64, error)
}

type Audit struct {
	repo   AuditRepo
	logger log.Logger
}

func NewAudit(repo AuditRepo, logger log.Logger) Audit {
	return Audit{
		repo:   repo,
		logger: logger,
	}
}

func (s Audit) SaveAuditAsync(ctx context.Context, userId int64, message string) {
	go func() {
		audit := entity.Audit{
			UserId:    int(userId),
			Message:   message,
			CreatedAt: time.Now().UTC(),
		}
		_, err := s.repo.Insert(context.Background(), audit)
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
		tokens, err = s.repo.All(ctx, limit, offset)
		if err != nil {
			return errors.WithMessage(err, "get all tokens")
		}
		return nil
	})
	group.Go(func() error {
		total, err = s.repo.Count(ctx)
		if err != nil {
			return errors.WithMessage(err, "count all tokens")
		}
		return nil
	})
	err = group.Wait()
	if err != nil {
		return nil, errors.WithMessage(err, "wait workers")
	}

	items := make([]domain.Audit, 0)
	for _, token := range tokens {
		items = append(items, domain.Audit(token))
	}
	result := domain.AuditResponse{
		TotalCount: int(total),
		Items:      items,
	}

	return &result, nil
}
