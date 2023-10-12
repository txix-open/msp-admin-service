package repository

import (
	"context"

	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/db/query"
	"github.com/integration-system/isp-kit/metrics/sql_metrics"
	"github.com/pkg/errors"
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
	sql_metrics.OperationLabelToContext(ctx, "Audit.Insert")

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
	sql_metrics.OperationLabelToContext(ctx, "Audit.All")

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
	sql_metrics.OperationLabelToContext(ctx, "Audit.Count")

	count := int64(0)
	err := r.db.SelectRow(ctx, &count, "select count(*) from audit")
	if err != nil {
		return 0, errors.WithMessage(err, "select count")
	}
	return count, nil
}
