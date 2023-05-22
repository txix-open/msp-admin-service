package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/db/query"
	"github.com/pkg/errors"
	"msp-admin-service/entity"
)

type UserRole struct {
	db db.DB
}

func NewUserRole(db db.DB) UserRole {
	return UserRole{db: db}
}

func (u UserRole) GetRolesByUserId(ctx context.Context, identity int) ([]int, error) {
	rolesQ, args, err := query.New().Select("role_id").
		From("user_roles").Where(squirrel.Eq{"user_id": identity}).ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}
	var roles []int
	err = u.db.Select(ctx, &roles, rolesQ, args...)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (u UserRole) GetRolesByUserIds(ctx context.Context, identity []int) ([]entity.UserRole, error) {
	rolesQ, args, err := query.New().Select("role_id", "user_id").
		From("user_roles").Where(squirrel.Eq{"user_id": identity}).ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	var roles []entity.UserRole
	err = u.db.Select(ctx, &roles, rolesQ, args...)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (u UserRole) Insert(ctx context.Context, id int, roleIds []int) error {
	rolesQ := query.New().
		Insert("user_roles").
		Columns("user_id", "role_id")

	for _, roleId := range roleIds {
		rolesQ = rolesQ.Values(id, roleId)
	}
	rolesQ = rolesQ.Suffix("ON CONFLICT DO NOTHING")
	rolesQResult, args, err := rolesQ.ToSql()
	if err != nil {
		return errors.WithMessage(err, "build query")
	}

	_, err = u.db.Exec(ctx, rolesQResult, args...)
	if err != nil {
		return errors.WithMessage(err, "exec")
	}

	return nil
}

func (u UserRole) ForceUpsert(ctx context.Context, id int, roleIds []int) error {
	deleteQ, args, err := query.New().
		Delete("user_roles").Where(squirrel.Eq{"user_id": id}).ToSql()
	if err != nil {
		return errors.WithMessage(err, "build query")
	}

	_, err = u.db.Exec(ctx, deleteQ, args...)
	if err != nil {
		return errors.WithMessage(err, "exec")
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
		return errors.WithMessage(err, "exec")
	}

	return nil
}
