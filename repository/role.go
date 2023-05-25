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
	sql_metrics.OperationLabelToContext(ctx, "Role.GetRoleByIds")

	q, args, err := query.New().
		Select("id, name, external_group, permissions, created_at, updated_at").
		From("roles").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	var roles []entity.Role
	err = r.db.Select(ctx, &roles, q, args...)

	switch {
	case err != nil:
		return nil, errors.WithMessagef(err, "db select: %s", q)
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
		return nil, errors.WithMessagef(err, "db select: %s", q)
	default:
		return &role, nil
	}
}

func (r Role) GetRoleByExternalGroup(ctx context.Context, group string) (*entity.Role, error) {
	sql_metrics.OperationLabelToContext(ctx, "Role.GetRoleByExternalGroup")

	q, args, err := query.New().
		Select("*").
		From("roles").
		Where(squirrel.Eq{"external_group": group}).
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
		return nil, errors.WithMessagef(err, "db select: %s", q)
	default:
		return &role, nil
	}
}

func (r Role) All(ctx context.Context) ([]entity.Role, error) {
	sql_metrics.OperationLabelToContext(ctx, "Role.All")

	roles := make([]entity.Role, 0)
	err := r.db.Select(ctx, &roles, "select id, name, external_group, permissions, created_at, updated_at from roles order by created_at")
	if err != nil {
		return nil, errors.WithMessage(err, "select all roles")
	}
	return roles, nil
}

func (r Role) InsertRole(ctx context.Context, role entity.Role) (*entity.Role, error) {
	q, args, err := query.New().Insert("roles").
		Columns("name", "permissions", "external_group").
		Values(role.Name, role.Permissions, role.ExternalGroup).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	var result entity.Role
	err = r.db.SelectRow(ctx, &result, q, args...)
	if err != nil {
		return nil, errors.WithMessagef(err, "insert: %s", q)
	}

	return &result, nil
}

func (r Role) Update(ctx context.Context, role entity.Role) (*entity.Role, error) {
	q, args, err := query.New().Update("roles").
		Set("name", role.Name).
		Set("permissions", role.Permissions).
		Set("external_group", role.ExternalGroup).
		Where(squirrel.Eq{"id": role.Id}).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	var result entity.Role
	err = r.db.SelectRow(ctx, &result, q, args...)
	if err != nil {
		return nil, errors.WithMessagef(err, "update: %s", q)
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
