package repository

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/db/query"
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
