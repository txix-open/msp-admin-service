package controller

import (
	"context"
	"strconv"

	"github.com/integration-system/isp-kit/grpc"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"msp-admin-service/domain"
)

type userService interface {
	GetProfileById(ctx context.Context, userId int64) (*domain.AdminUserShort, error)
	GetUsers(ctx context.Context, identities domain.UsersRequest) (*domain.UsersResponse, error)
	CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error)
	UpdateUser(ctx context.Context, req domain.UpdateUserRequest) (*domain.User, error)
	DeleteUsers(ctx context.Context, ids []int64) (int, error)
	Block(ctx context.Context, userId int) error
	Roles(ctx context.Context) ([]domain.Role, error)
	GetById(ctx context.Context, userId int) (*domain.User, error)
}

type User struct {
	userService userService
}

func NewUser(userService userService) User {
	return User{
		userService: userService,
	}
}

// GetProfile
// @Tags user
// @Summary Получить профиль
// @Description Получить данные профиля
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Success 200 {object} domain.AdminUserShort
// @Failure 400 {object} domain.GrpcError "Невалидный токен"
// @Failure 404 {object} domain.GrpcError "Пользователя не существует"
// @Failure 500 {object} domain.GrpcError
// @Router /user/get_profile [POST]
func (u User) GetProfile(ctx context.Context, authData grpc.AuthData) (*domain.AdminUserShort, error) {
	adminIdInString, err := grpc.StringFromMd(domain.AdminAuthIdHeader, metadata.MD(authData))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	adminId, err := strconv.Atoi(adminIdInString)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "admin id is not a number")
	}

	profile, err := u.userService.GetProfileById(ctx, int64(adminId))

	switch {
	case errors.Is(err, domain.ErrUnauthenticated):
		return nil, status.Error(codes.Unauthenticated, "user is blocked")
	case errors.Is(err, domain.ErrNotFound):
		return nil, status.Error(codes.NotFound, "user not found")
	case err != nil:
		return nil, errors.WithMessage(err, "get profile")
	default:
		return profile, nil
	}
}

// GetUsers
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
func (u User) GetUsers(ctx context.Context, identities domain.UsersRequest) (*domain.UsersResponse, error) {
	users, err := u.userService.GetUsers(ctx, identities)
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
func (u User) CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error) {
	user, err := u.userService.CreateUser(ctx, req)

	switch {
	case errors.Is(err, domain.ErrAlreadyExists):
		return nil, status.Error(codes.AlreadyExists, "user with the same email already exists")
	case err != nil:
		return nil, errors.WithMessage(err, "create user")
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
func (u User) UpdateUser(ctx context.Context, req domain.UpdateUserRequest) (*domain.User, error) {
	result, err := u.userService.UpdateUser(ctx, req)

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
func (u User) DeleteUser(ctx context.Context, identities domain.IdentitiesRequest) (*domain.DeleteResponse, error) {
	deletedCount, err := u.userService.DeleteUsers(ctx, identities.Ids)
	if err != nil {
		return nil, errors.WithMessage(err, "delete")
	}

	return &domain.DeleteResponse{Deleted: deletedCount}, nil
}

// Block
// @Tags user
// @Summary Метод блокировки/разблокировки пользователя
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.IdRequest true "Тело запроса"
// @Success 200
// @Failure 400 {object} domain.GrpcError "Невалидное тело запроса"
// @Failure 500 {object} domain.GrpcError
// @Router /user/block_user [POST]
func (u User) Block(ctx context.Context, identities domain.IdRequest) error {
	err := u.userService.Block(ctx, identities.UserId)
	if err != nil {
		return errors.WithMessage(err, "block")
	}
	return nil
}

// GetRoles
// @Tags user
// @Summary Список доступных ролей
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Success 200 {array} domain.Role
// @Failure 400 {object} domain.GrpcError
// @Failure 500 {object} domain.GrpcError
// @Router /user/get_roles [POST]
func (u User) GetRoles(ctx context.Context) ([]domain.Role, error) {
	users, err := u.userService.Roles(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get roles")
	}

	return users, nil
}

// GetById
// @Tags user
// @Summary Метод получения данных по пользователю
// @Accept json
// @Produce json
// @Param body body domain.IdRequest true "Тело запроса"
// @Success 200 {object} domain.User
// @Failure 400 {object} domain.GrpcError "Невалидное тело запроса"
// @Failure 500 {object} domain.GrpcError
// @Router /user/get_by_id [POST]
func (u User) GetById(ctx context.Context, identities domain.IdRequest) (*domain.User, error) {
	return u.userService.GetById(ctx, identities.UserId)
}
