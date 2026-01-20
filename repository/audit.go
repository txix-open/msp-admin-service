package repository

import (
	"context"
	"time"

	"msp-admin-service/domain"
	"msp-admin-service/entity"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/db/query"
	"github.com/txix-open/isp-kit/metrics/sql_metrics"
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

//nolint:dupl,gosec
func (r Audit) All(ctx context.Context, req domain.AuditPageRequest) ([]entity.Audit, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Audit.All")

	q := query.New().
		Select("*").
		From("audit").
		Offset(uint64(req.Offset)).
		Limit(uint64(req.Limit))

	if req.Order != nil {
		q = q.OrderBy(req.Order.Field + " " + req.Order.Type)
	}

	query, args, err := reqAuditQuery(q, req.Query).ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	logs := make([]entity.Audit, 0)
	err = r.db.Select(ctx, &logs, query, args...)
	if err != nil {
		return nil, errors.WithMessagef(err, "select query: %s", query)
	}

	return logs, nil
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

func reqAuditQuery(q squirrel.SelectBuilder, reqQuery *domain.AuditQuery) squirrel.SelectBuilder {
	if reqQuery == nil {
		return q
	}

	switch {
	case reqQuery.Id != nil:
		{
			q = q.Where("id = ?", *reqQuery.Id)
		}
	case reqQuery.UserId != nil:
		{
			q = q.Where("user_id = ?", *reqQuery.UserId)
		}
	case reqQuery.Message != nil:
		{
			q = q.Where("message LIKE ?", "%"+*reqQuery.Message+"%")
		}
	case reqQuery.CreatedAt != nil:
		{
			q = q.Where(squirrel.GtOrEq{"created_at": reqQuery.CreatedAt.From},
				squirrel.LtOrEq{"created_at": reqQuery.CreatedAt.To})
		}
	}

	return q
}
