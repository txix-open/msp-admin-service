// nolint: gosec
package repository

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/db/query"
	"github.com/txix-open/isp-kit/metrics/sql_metrics"
	"msp-admin-service/entity"
)

type Audit struct {
	db db.DB
}

func NewAudit(db db.DB) Audit {
	return Audit{
		db: db,
	}
}

func (r Audit) Insert(ctx context.Context, log entity.Audit) (int, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Audit.Insert")

	query, args, err := query.New().
		Insert("audit").
		Columns("user_id", "message", "event", "created_at").
		Values(log.UserId, log.Message, log.Event, log.CreatedAt).
		Suffix("returning id").
		ToSql()
	if err != nil {
		return 0, errors.WithMessage(err, "build query")
	}

	id := 0
	err = r.db.SelectRow(ctx, &id, query, args...)
	if err != nil {
		return 0, errors.WithMessagef(err, "select row: %s", query)
	}

	return id, nil
}

func (r Audit) All(ctx context.Context, limit int, offset int) ([]entity.Audit, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Audit.All")

	query, args, err := query.New().
		Select("*").
		From("audit").
		OrderBy("created_at DESC").
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	tokens := make([]entity.Audit, 0)
	err = r.db.Select(ctx, &tokens, query, args...)
	if err != nil {
		return nil, errors.WithMessagef(err, "select query: %s", query)
	}

	return tokens, nil
}

func (r Audit) Count(ctx context.Context) (int64, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Audit.Count")

	count := int64(0)
	err := r.db.SelectRow(ctx, &count, "select count(*) from audit")
	if err != nil {
		return 0, errors.WithMessage(err, "select count")
	}
	return count, nil
}

func (r Audit) DeleteUpToCreatedAt(ctx context.Context, createdAt time.Time) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Audit.DeleteUpToCreatedAt")

	q, args, err := query.New().
		Delete("audit").
		Where("created_at < ?", createdAt).
		ToSql()
	if err != nil {
		return errors.WithMessage(err, "build query")
	}

	_, err = r.db.Exec(ctx, q, args...)
	if err != nil {
		return errors.WithMessagef(err, "delete audit")
	}

	return nil
}
