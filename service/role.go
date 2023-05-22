package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
)

type roleRoleRepo interface {
	All(ctx context.Context) ([]entity.Role, error)
	Insert(ctx context.Context, role entity.Role) (*entity.Role, error)
	Update(ctx context.Context, role entity.Role) (*entity.Role, error)
	Delete(ctx context.Context, id int) error
}

type Role struct {
	roleRepo roleRoleRepo
}

func NewRole(roleRepo roleRoleRepo) Role {
	return Role{
		roleRepo: roleRepo,
	}
}

func (u Role) All(ctx context.Context) ([]domain.Role, error) {
	roles, err := u.roleRepo.All(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get all roles")
	}

	resRoles := make([]domain.Role, 0)
	for _, role := range roles {
		resRoles = append(resRoles, u.toDomain(role))
	}

	return resRoles, nil
}

func (u Role) Create(ctx context.Context, req domain.CreateRoleRequest, adminId int64) (*domain.Role, error) {
	role, err := u.roleRepo.Insert(ctx, entity.Role{
		Name:          req.Name,
		ExternalGroup: req.ExternalGroup,
		ChangeMessage: req.ChangeMessage,
		Permissions:   req.Permissions,
	})

	if err != nil {
		return nil, errors.WithMessage(err, "create user")
	}

	u.auditService.SaveAuditAsync(ctx, adminId, fmt.Sprintf("Роль. Создание новой роли %s. %s",
		role.Name,
		role.ChangeMessage,
	))

	result := u.toDomain(*role)
	return &result, nil
}

func (u Role) Update(ctx context.Context, req domain.UpdateRoleRequest, adminId int64) (*domain.Role, error) {
	role, err := u.roleRepo.Update(ctx, entity.Role{
		Id:            req.Id,
		Name:          req.Name,
		ExternalGroup: req.ExternalGroup,
		ChangeMessage: req.ChangeMessage,
		Permissions:   req.Permissions,
	})

	if err != nil {
		return nil, errors.WithMessage(err, "update user")
	}

	u.auditService.SaveAuditAsync(ctx, adminId, fmt.Sprintf("Роль. Изменение роли %s. %s",
		role.Name,
		role.ChangeMessage,
	))
	result := u.toDomain(*role)
	return &result, nil
}

func (u Role) Delete(ctx context.Context, req domain.DeleteRoleRequest, adminId int64) error {
	err := u.roleRepo.Delete(ctx, req.Id)
	if err != nil {
		return errors.WithMessage(err, "delete role")
	}

	u.auditService.SaveAuditAsync(ctx, adminId, fmt.Sprintf("Роль. Удаление роли. ID: %d", req.Id))
	return nil
}

func (u Role) toDomain(role entity.Role) domain.Role {
	return domain.Role{
		Id:            role.Id,
		Name:          role.Name,
		ExternalGroup: role.ExternalGroup,
		ChangeMessage: role.ChangeMessage,
		Permissions:   role.Permissions,
		CreatedAt:     role.CreatedAt,
		UpdatedAt:     role.UpdatedAt,
	}
}
