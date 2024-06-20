package controller

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/log"
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
	adminId, err := getAdminId(authData)
	if err != nil {
		return err
	}

	err = a.authService.Logout(ctx, adminId)
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
		return nil, status.Error(codes.InvalidArgument, "plain auth is not available")
	case errors.Is(err, domain.ErrUnauthenticated):
		a.logger.Error(ctx, err.Error())
		return nil, status.Error(codes.Unauthenticated, "invalid credential")
	case errors.Is(err, domain.ErrTooManyLoginRequests):
		return nil, status.Error(codes.ResourceExhausted, "too many requests")
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
		return nil, status.Error(codes.FailedPrecondition, "sudir auth is not configured")
	case errors.Is(err, domain.ErrUnauthenticated):
		return nil, status.Error(codes.Unauthenticated, "invalid code")
	case err != nil:
		return nil, errors.WithMessage(err, "login with sudir")
	default:
		return auth, nil
	}
}

func getAdminId(authData grpc.AuthData) (int64, error) {
	token, err := grpc.StringFromMd(domain.AdminAuthIdHeader, metadata.MD(authData))
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, err.Error())
	}

	adminId, err := strconv.Atoi(token)
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, "admin id is not a number")
	}

	return int64(adminId), nil
}
