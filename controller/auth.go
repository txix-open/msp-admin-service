package controller

import (
	"context"

	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"msp-admin-service/domain"
)

type authService interface {
	Login(ctx context.Context, authRequest domain.AuthRequest) (*domain.Auth, error)
	LoginWithSudir(ctx context.Context, request domain.SudirAuthRequest) (*domain.Auth, error)
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
func (a Auth) Logout() error {
	return nil
}

// Login
// @Tags auth
// @Summary Авторизация по логину и паролю
// @Description Авторизация с получением токена администратора
// @Accept json
// @Produce json
// @Param body body domain.AuthRequest true "Тело запроса"
// @Success 200 {object} domain.Auth
// @Failure 400 {object} domain.GrpcError
// @Failure 401 {object} domain.GrpcError "Данные для авторизации не верны"
// @Failure 500 {object} domain.GrpcError
// @Router /auth/login [POST]
func (a Auth) Login(ctx context.Context, authRequest domain.AuthRequest) (*domain.Auth, error) {
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
// @Param body body domain.SudirAuthRequest true "Тело запроса"
// @Success 200 {object} domain.Auth
// @Failure 401 {object} domain.GrpcError "Некорректный код для авторизации"
// @Failure 412 {object} domain.GrpcError "Авторизация СУДИР не настроена на сервере"
// @Failure 500 {object} domain.GrpcError
// @Router /auth/login_with_sudir [POST]
func (a Auth) LoginWithSudir(ctx context.Context, request domain.SudirAuthRequest) (*domain.Auth, error) {
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
