package repository

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/db/query"
	"github.com/integration-system/isp-kit/metrics/sql_metrics"
	"github.com/pkg/errors"
	"msp-admin-service/entity"
)

type AuditEvent struct {
	db db.DB
}

func NewAuditEvent(db db.DB) AuditEvent {
	return AuditEvent{
		db: db,
	}
}

func (r AuditEvent) All(ctx context.Context) ([]entity.AuditEvent, error) {
	sql_metrics.OperationLabelToContext(ctx, "Audit_Event.All")

	q, args, err := query.New().
		Select("*").
		From("audit_event").
		OrderBy("enable DESC").
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	tokens := make([]entity.AuditEvent, 0)
	err = r.db.Select(ctx, &tokens, q, args...)
	if err != nil {
		return nil, errors.WithMessagef(err, "select query: %s", q)
	}

	return tokens, nil
}

func (r AuditEvent) Upsert(ctx context.Context, eventList []entity.AuditEvent) error {
	sql_metrics.OperationLabelToContext(ctx, "Audit_Event.Upsert")

	qBuilder := query.New().
		Insert("audit_event").
		Columns("event", "enable")
	for _, event := range eventList {
		qBuilder = qBuilder.Values(event.Event, event.Enable)
	}
	q, args, err := qBuilder.
		Suffix("ON CONFLICT (event) DO UPDATE SET enable = EXCLUDED.enable").
		ToSql()
	if err != nil {
		return errors.WithMessage(err, "build query")
	}

	_, err = r.db.Exec(ctx, q, args...)
	if err != nil {
		return errors.WithMessage(err, "upsert audit_event list")
	}

	return nil
}

func (r AuditEvent) IsEnable(ctx context.Context, event string) (bool, error) {
	q, args, err := query.New().
		Select("enable").
		From("audit_event").
		Where(squirrel.Eq{"event": event}).
		ToSql()
	if err != nil {
		return false, errors.WithMessage(err, "build query")
	}

	result := false
	err = r.db.SelectRow(ctx, &result, q, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, errors.WithMessagef(err, "select event")
	}

	return result, nil
}
