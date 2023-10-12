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
	InsertRole(ctx context.Context, role entity.Role) (*entity.Role, error)
	Update(ctx context.Context, role entity.Role) (*entity.Role, error)
	Delete(ctx context.Context, id int) error
	GetRoleByName(ctx context.Context, name string) (*entity.Role, error)
	GetRoleByIds(ctx context.Context, id []int) ([]entity.Role, error)
}

type Role struct {
	roleRepo     roleRoleRepo
	auditService auditService
}

func NewRole(roleRepo roleRoleRepo, audit auditService) Role {
	return Role{
		roleRepo:     roleRepo,
		auditService: audit,
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
	role, err := u.roleRepo.GetRoleByName(ctx, req.Name)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		break
	case err != nil:
		return nil, errors.WithMessagef(err, "get role by name")
	case role != nil:
		return nil, domain.ErrAlreadyExists
	}

	role, err = u.roleRepo.InsertRole(ctx, entity.Role{
		Name:          req.Name,
		ExternalGroup: req.ExternalGroup,
		Permissions:   req.Permissions,
	})

	if err != nil {
		return nil, errors.WithMessage(err, "create role")
	}

	u.auditService.SaveAuditAsync(ctx, adminId,
		fmt.Sprintf("Роль. Создание новой роли %s. Причина: %s", req.Name, req.ChangeMessage),
		entity.EventRoleChanged,
	)

	result := u.toDomain(*role)
	return &result, nil
}

func (u Role) Update(ctx context.Context, req domain.UpdateRoleRequest, adminId int64) (*domain.Role, error) {
	roleByName, err := u.roleRepo.GetRoleByName(ctx, req.Name)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		break
	case err != nil:
		return nil, errors.WithMessagef(err, "get role by name")
	}

	roles, err := u.roleRepo.GetRoleByIds(ctx, []int{req.Id})
	switch {
	case err != nil:
		return nil, errors.WithMessagef(err, "get role by id")
	case len(roles) == 0:
		return nil, domain.ErrNotFound
	case roleByName != nil && roleByName.Id != roles[0].Id:
		return nil, domain.ErrAlreadyExists
	}

	role, err := u.roleRepo.Update(ctx, entity.Role{
		Id:            req.Id,
		Name:          req.Name,
		ExternalGroup: req.ExternalGroup,
		Permissions:   req.Permissions,
	})

	if err != nil {
		return nil, errors.WithMessage(err, "update role")
	}

	u.auditService.SaveAuditAsync(ctx, adminId,
		fmt.Sprintf("Роль. Изменение роли %s. Причина: %s", req.Name, req.ChangeMessage),
		entity.EventRoleChanged,
	)
	result := u.toDomain(*role)
	return &result, nil
}

func (u Role) Delete(ctx context.Context, req domain.DeleteRoleRequest, adminId int64) error {
	err := u.roleRepo.Delete(ctx, req.Id)
	if err != nil {
		return errors.WithMessage(err, "delete role")
	}

	u.auditService.SaveAuditAsync(ctx, adminId, fmt.Sprintf("Роль. Удаление роли. ID: %d", req.Id),
		entity.EventRoleChanged,
	)
	return nil
}

func (u Role) toDomain(role entity.Role) domain.Role {
	return domain.Role{
		Id:            role.Id,
		Name:          role.Name,
		ExternalGroup: role.ExternalGroup,
		Permissions:   role.Permissions,
		Immutable:     role.Immutable,
		Exclusive:     role.Exclusive,
		CreatedAt:     role.CreatedAt,
		UpdatedAt:     role.UpdatedAt,
	}
}
