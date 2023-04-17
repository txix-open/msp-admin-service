package repository

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/db/query"
	"github.com/integration-system/isp-kit/metrics/sql_metrics"
	"github.com/pkg/errors"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
)

type Role struct {
	db db.DB
}

func NewRole(db db.DB) Role {
	return Role{db: db}
}

func (r Role) GetRoleById(ctx context.Context, id int) (*entity.Role, error) {
	sql_metrics.OperationLabelToContext(ctx, "Role.GetRoleById")

	q, args, err := query.New().
		Select("*").
		From("roles").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	role := entity.Role{}
	err = r.db.SelectRow(ctx, &role, q, args...)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, domain.ErrNotFound
	case err != nil:
		return nil, errors.WithMessage(err, "db select")
	default:
		return &role, nil
	}
}

func (r Role) GetRoleByName(ctx context.Context, name string) (*entity.Role, error) {
	sql_metrics.OperationLabelToContext(ctx, "Role.GetRoleByName")

	q, args, err := query.New().
		Select("*").
		From("roles").
		Where(squirrel.Eq{"name": name}).
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	role := entity.Role{}
	err = r.db.SelectRow(ctx, &role, q, args...)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, domain.ErrNotFound
	case err != nil:
		return nil, errors.WithMessage(err, "db select")
	default:
		return &role, nil
	}
}

func (r Role) UpsertRoleByName(ctx context.Context, role entity.Role) (int, error) {
	sql_metrics.OperationLabelToContext(ctx, "Role.UpsertRoleByName")

	query := `
	insert into roles (name, rights, description) 
	values ($1, $2, $3) on conflict (name) do update 
    set name = excluded.name, rights = excluded.rights, description = excluded.description
	returning id
`
	id := 0
	err := r.db.SelectRow(ctx, &id, query, role.Name, role.Rights, role.Description)
	if err != nil {
		return 0, errors.WithMessagef(err, "select %s", query)
	}

	return id, nil
}

func (r Role) All(ctx context.Context) ([]entity.Role, error) {
	sql_metrics.OperationLabelToContext(ctx, "Role.All")

	roles := make([]entity.Role, 0)
	err := r.db.Select(ctx, &roles, "select * from roles order by created_at")
	if err != nil {
		return nil, errors.WithMessage(err, "select all roles")
	}
	return roles, nil
}
