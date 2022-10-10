package controller

import (
	"context"
	"strconv"

	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"msp-admin-service/domain"
)

type authService interface {
	Login(ctx context.Context, request domain.LoginRequest) (*domain.LoginResponse, error)
	LoginWithSudir(ctx context.Context, request domain.LoginSudirRequest) (*domain.LoginResponse, error)
	Logout(ctx context.Context, adminId int64) error
}

type Auth struct {
	authService authService
	logger      log.Logger
}

func NewAuth(authService authService, logger log.Logger) Auth {
	return Auth{
		authService: authService,
		logger:      logger,
	}
}

// Logout
// @Tags auth
// @Summary Выход из авторизованной сессии
// @Description Выход из авторизованной сессии администрирования
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Success 200
// @Failure 400 {object} domain.GrpcError "Невалидный токен"
// @Failure 500 {object} domain.GrpcError
// @Router /auth/logout [POST]
func (a Auth) Logout(ctx context.Context, authData grpc.AuthData) error {
	token, err := grpc.StringFromMd(domain.AdminAuthIdHeader, metadata.MD(authData))
	if err != nil {
		return status.Error(codes.InvalidArgument, "Отсутствует идентификатор")
	}

	adminId, err := strconv.Atoi(token)
	if err != nil {
		return status.Error(codes.InvalidArgument, "Недействительный идентификатор")
	}

	err = a.authService.Logout(ctx, int64(adminId))
	if err != nil {
		return errors.WithMessage(err, "logout")
	}

	return nil
}

// Login
// @Tags auth
// @Summary Авторизация по логину и паролю
// @Description Авторизация с получением токена администратора
// @Accept json
// @Produce json
// @Param body body domain.LoginRequest true "Тело запроса"
// @Success 200 {object} domain.LoginResponse
// @Failure 400 {object} domain.GrpcError
// @Failure 401 {object} domain.GrpcError "Данные для авторизации не верны"
// @Failure 500 {object} domain.GrpcError
// @Router /auth/login [POST]
func (a Auth) Login(ctx context.Context, authRequest domain.LoginRequest) (*domain.LoginResponse, error) {
	auth, err := a.authService.Login(ctx, authRequest)

	switch {
	case errors.Is(err, domain.ErrSudirAuthorization):
		return nil, status.Error(codes.InvalidArgument, "Пользователь имеет только авторизацию СУДИР")
	case errors.Is(err, domain.ErrUnauthenticated):
		a.logger.Error(ctx, err.Error())
		return nil, status.Error(codes.Unauthenticated, "Данные для авторизации не верны")
	case err != nil:
		return nil, errors.WithMessage(err, "login")
	default:
		return auth, nil
	}
}

// LoginWithSudir
// @Tags auth
// @Summary Авторизация по авторизационному коду от СУДИР
// @Description Авторизация с получением токена администратора
// @Accept json
// @Produce json
// @Param body body domain.LoginSudirRequest true "Тело запроса"
// @Success 200 {object} domain.LoginResponse
// @Failure 401 {object} domain.GrpcError "Некорректный код для авторизации"
// @Failure 412 {object} domain.GrpcError "Авторизация СУДИР не настроена на сервере"
// @Failure 500 {object} domain.GrpcError
// @Router /auth/login_with_sudir [POST]
func (a Auth) LoginWithSudir(ctx context.Context, request domain.LoginSudirRequest) (*domain.LoginResponse, error) {
	auth, err := a.authService.LoginWithSudir(ctx, request)

	switch {
	case errors.Is(err, domain.ErrSudirAuthIsMissed):
		return nil, status.Error(codes.FailedPrecondition, "Авторизация СУДИР не настроена на сервере")
	case errors.Is(err, domain.ErrUnauthenticated):
		return nil, status.Error(codes.Unauthenticated, "Некорректный код для авторизации")
	case err != nil:
		return nil, errors.WithMessage(err, "login with sudir")
	default:
		return auth, nil
	}
}
