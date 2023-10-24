package controller

import (
	"context"

	"github.com/integration-system/isp-kit/grpc"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"msp-admin-service/domain"
)

type roleService interface {
	All(ctx context.Context) ([]domain.Role, error)
	Create(ctx context.Context, req domain.CreateRoleRequest, adminId int64) (*domain.Role, error)
	Update(ctx context.Context, req domain.UpdateRoleRequest, adminId int64) (*domain.Role, error)
	Delete(ctx context.Context, req domain.DeleteRoleRequest, adminId int64) error
}

type Role struct {
	roleService roleService
}

func NewRole(roleService roleService) Role {
	return Role{
		roleService: roleService,
	}
}

// All
// @Tags role
// @Summary Список ролей
// @Description Получить список ролей
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Success 200 {array} domain.Role
// @Failure 400 {object} domain.GrpcError
// @Failure 500 {object} domain.GrpcError
// @Router /role/all [POST]
func (u Role) All(ctx context.Context) ([]domain.Role, error) {
	users, err := u.roleService.All(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get roles")
	}

	return users, nil
}

// CreateRole
// @Tags user
// @Summary Создать роль
// @Description Создать роль
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.CreateRoleRequest true "Тело запроса"
// @Success 200 {object} domain.User
// @Failure 400 {object} domain.GrpcError "Невалидное тело запроса"
// @Failure 409 {object} domain.GrpcError "Роль с указанным именем уже существует"
// @Failure 500 {object} domain.GrpcError
// @Router /role/create [POST]
func (u Role) CreateRole(ctx context.Context, authData grpc.AuthData, req domain.CreateRoleRequest) (*domain.Role, error) {
	adminId, err := getAdminId(authData)
	if err != nil {
		return nil, err
	}

	role, err := u.roleService.Create(ctx, req, adminId)
	switch {
	case errors.Is(err, domain.ErrAlreadyExists):
		return nil, status.Error(codes.AlreadyExists, "role with current name already exists")
	case err != nil:
		return nil, errors.WithMessage(err, "create role")
	default:
		return role, nil
	}
}

// UpdateRole
// @Tags role
// @Summary Обновить роль
// @Description Обновить данные существующую роль
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.UpdateRoleRequest true "Тело запроса"
// @Success 200 {object} domain.Role
// @Failure 404 {object} domain.GrpcError "Роль с указанным id не существует"
// @Failure 409 {object} domain.GrpcError "Роль с указанным именем уже существует"
// @Failure 500 {object} domain.GrpcError
// @Router /role/update [POST]
func (u Role) UpdateRole(ctx context.Context, authData grpc.AuthData, req domain.UpdateRoleRequest) (*domain.Role, error) {
	adminId, err := getAdminId(authData)
	if err != nil {
		return nil, err
	}

	result, err := u.roleService.Update(ctx, req, adminId)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return nil, status.Error(codes.NotFound, "role not found")
	case errors.Is(err, domain.ErrAlreadyExists):
		return nil, status.Error(codes.AlreadyExists, "role with current name already exists")
	case err != nil:
		return nil, errors.WithMessage(err, "update role")
	default:
		return result, nil
	}
}

// DeleteRole
// @Tags role
// @Summary Удалить роль
// @Description Удалить существующую роль
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.DeleteRoleRequest true "Тело запроса"
// @Success 200
// @Failure 400 {object} domain.GrpcError "Невалидное тело запроса"
// @Failure 500 {object} domain.GrpcError
// @Router /role/delete [POST]
func (u Role) DeleteRole(ctx context.Context, authData grpc.AuthData, req domain.DeleteRoleRequest) error {
	adminId, err := getAdminId(authData)
	if err != nil {
		return err
	}

	err = u.roleService.Delete(ctx, req, adminId)
	if err != nil {
		return errors.WithMessage(err, "delete")
	}

	return nil
}
