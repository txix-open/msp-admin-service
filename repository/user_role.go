package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/db/query"
	"github.com/txix-open/isp-kit/metrics/sql_metrics"
	"msp-admin-service/entity"
)

type UserRole struct {
	db db.DB
}

func NewUserRole(db db.DB) UserRole {
	return UserRole{db: db}
}

func (u UserRole) GetRolesByUserIds(ctx context.Context, identity []int) ([]entity.UserRole, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "UserRole.GetRolesByUserIds")

	rolesQ, args, err := query.New().Select("role_id", "user_id").
		From("user_roles").Where(squirrel.Eq{"user_id": identity}).ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	var roles []entity.UserRole
	err = u.db.Select(ctx, &roles, rolesQ, args...)
	switch {
	case err != nil:
		return nil, errors.WithMessagef(err, "db select: %s", rolesQ)
	default:
		return roles, nil
	}
}

func (u UserRole) GetRoleEntitiesByUserId(ctx context.Context, userId int) ([]entity.Role, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "UserRole.GetRoleEntitiesByUserId")

	query, args, err := query.New().
		Select("r.*").
		From("user_roles ur").
		Join("roles r on ur.role_id = r.id").
		Where(squirrel.Eq{
			"ur.user_id": userId,
		}).OrderBy("r.id").
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	result := make([]entity.Role, 0)
	err = u.db.Select(ctx, &result, query, args...)
	if err != nil {
		return nil, errors.WithMessagef(err, "db select: %s", query)
	}

	return result, nil
}

func (u UserRole) UpsertUserRoleLinks(ctx context.Context, id int, roleIds []int) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "UserRole.UpsertUserRoleLinks")

	deleteQ, args, err := query.New().
		Delete("user_roles").Where(squirrel.Eq{"user_id": id}).ToSql()
	if err != nil {
		return errors.WithMessage(err, "build query")
	}

	_, err = u.db.Exec(ctx, deleteQ, args...)
	if err != nil {
		return errors.WithMessagef(err, "exec: %s", deleteQ)
	}

	if len(roleIds) == 0 {
		return nil
	}

	rolesQ := query.New().
		Insert("user_roles").
		Columns("user_id", "role_id")
	for _, roleId := range roleIds {
		rolesQ = rolesQ.Values(id, roleId)
	}
	rolesQResult, args, err := rolesQ.ToSql()
	if err != nil {
		return errors.WithMessage(err, "build query")
	}

	_, err = u.db.Exec(ctx, rolesQResult, args...)
	if err != nil {
		return errors.WithMessagef(err, "exec: %s", rolesQResult)
	}

	return nil
}
