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

func (r Role) GetRoleByIds(ctx context.Context, id []int) ([]entity.Role, error) {
	sql_metrics.OperationLabelToContext(ctx, "Role.GetRoleById")

	q, args, err := query.New().
		Select("id, name, change_message, external_group, permissions::jsonb, created_at, updated_at").
		From("roles").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	var roles []entity.Role
	err = r.db.Select(ctx, &roles, q, args...)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, domain.ErrNotFound
	case err != nil:
		return nil, errors.WithMessage(err, "db select")
	default:
		return roles, nil
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
	insert into roles (name, permissions, external_group) 
	values ($1, $2, $3) on conflict (name) do update 
    set name = excluded.name, external_group = excluded.external_group
	returning id
`
	id := 0
	err := r.db.SelectRow(ctx, &id, query, role.Name, role.Permissions, role.ExternalGroup)
	if err != nil {
		return 0, errors.WithMessagef(err, "select %s", query)
	}

	return id, nil
}

func (r Role) All(ctx context.Context) ([]entity.Role, error) {
	sql_metrics.OperationLabelToContext(ctx, "Role.All")

	roles := make([]entity.Role, 0)
	err := r.db.Select(ctx, &roles, "select id, name, external_group, change_message, permissions::jsonb, created_at, updated_at from roles order by created_at")
	if err != nil {
		return nil, errors.WithMessage(err, "select all roles")
	}
	return roles, nil
}

func (r Role) InsertRole(ctx context.Context, role entity.Role) (*entity.Role, error) {
	q, args, err := query.New().Insert("roles").
		Columns("name", "permissions", "change_message").
		Values(role.Name, role.Permissions, role.ChangeMessage).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	var result entity.Role
	err = r.db.SelectRow(ctx, &result, q, args...)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r Role) Update(ctx context.Context, role entity.Role) (*entity.Role, error) {
	q, args, err := query.New().Update("roles").
		Set("name", role.Name).
		Set("permissions", role.Permissions).
		Set("change_message", role.ChangeMessage).
		Where(squirrel.Eq{"id": role.Id}).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	var result entity.Role
	err = r.db.SelectRow(ctx, &result, q, args...)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r Role) Delete(ctx context.Context, id int) error {
	q, args, err := query.New().Delete("roles").Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return errors.WithMessage(err, "build query")
	}

	_, err = r.db.Exec(ctx, q, args...)
	if err != nil {
		return errors.WithMessage(err, "delete role")
	}

	return nil
}
