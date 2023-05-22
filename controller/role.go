package controller

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"msp-admin-service/domain"
)

type roleService interface {
	All(ctx context.Context) ([]domain.Role, error)
	Create(ctx context.Context, req domain.CreateRoleRequest) (*domain.Role, error)
	Update(ctx context.Context, req domain.UpdateRoleRequest) (*domain.Role, error)
	Delete(ctx context.Context, req domain.DeleteRoleRequest) error
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
// @Tags user
// @Summary Список пользователей
// @Description Получить список пользователей
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.UsersRequest true "Тело запроса"
// @Success 200 {object} domain.UsersResponse
// @Failure 400 {object} domain.GrpcError
// @Failure 500 {object} domain.GrpcError
// @Router /user/get_users [POST]
func (u Role) All(ctx context.Context) ([]domain.Role, error) {
	users, err := u.roleService.All(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get users")
	}

	return users, nil
}

// CreateUser
// @Tags user
// @Summary Создать пользователя
// @Description Создать пользователя
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.CreateUserRequest true "Тело запроса"
// @Success 200 {object} domain.User
// @Failure 400 {object} domain.GrpcError "Невалидное тело запроса"
// @Failure 409 {object} domain.GrpcError "Пользователь с указанным email уже существует"
// @Failure 500 {object} domain.GrpcError
// @Router /user/create_user [POST]
func (u Role) CreateRole(ctx context.Context, req domain.CreateRoleRequest) (*domain.Role, error) {
	user, err := u.roleService.Create(ctx, req)

	switch {
	case errors.Is(err, domain.ErrAlreadyExists):
		return nil, status.Error(codes.AlreadyExists, "role with the same user-role pair")
	case err != nil:
		return nil, errors.WithMessage(err, "create role")
	default:
		return user, nil
	}
}

// UpdateUser
// @Tags user
// @Summary Обновить пользователя
// @Description Обновить данные существующего пользователя
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.UpdateUserRequest true "Тело запроса"
// @Success 200 {object} domain.User
// @Failure 400 {object} domain.GrpcError "Невалидное тело запроса"
// @Failure 404 {object} domain.GrpcError "Пользователь с указанным id не существует"
// @Failure 409 {object} domain.GrpcError "Пользователь с указанным email уже существует"
// @Failure 500 {object} domain.GrpcError
// @Router /user/update_user [POST]
func (u Role) UpdateRole(ctx context.Context, req domain.UpdateRoleRequest) (*domain.Role, error) {
	result, err := u.roleService.Update(ctx, req)

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return nil, status.Error(codes.NotFound, "user not found")
	case errors.Is(err, domain.ErrInvalid):
		return nil, status.Error(codes.InvalidArgument, "user modification is not available")
	case errors.Is(err, domain.ErrAlreadyExists):
		return nil, status.Error(codes.AlreadyExists, "user with the same email already exists")
	case err != nil:
		return nil, errors.WithMessage(err, "update user")
	default:
		return result, nil
	}
}

// DeleteUser
// @Tags user
// @Summary Удалить пользователя
// @Description Удалить существующего пользователя
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.IdentitiesRequest true "Тело запроса"
// @Success 200 {object} domain.DeleteResponse
// @Failure 400 {object} domain.GrpcError "Невалидное тело запроса"
// @Failure 500 {object} domain.GrpcError
// @Router /user/delete_user [POST]
func (u Role) DeleteRole(ctx context.Context, req domain.DeleteRoleRequest) error {
	err := u.roleService.Delete(ctx, req)
	if err != nil {
		return errors.WithMessage(err, "delete")
	}

	return nil
}
