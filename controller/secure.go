package controller

import (
	"context"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc/apierrors"
	"msp-admin-service/domain"
)

type SecureService interface {
	Authenticate(ctx context.Context, token string) (int64, error)
	Authorize(ctx context.Context, adminId int, permission string) (bool, error)
}

type Secure struct {
	service SecureService
}

func NewSecure(service SecureService) Secure {
	return Secure{
		service: service,
	}
}

// Authenticate
// @Tags secure
// @Summary Метод аутентификации токена
// @Description Проверяет токен и возвращает идентификатор администратора
// @Accept json
// @Produce json
// @Param body body domain.SecureAuthRequest true "Тело запроса"
// @Success 200 {object} domain.SecureAuthResponse
// @Failure 500 {object} domain.GrpcError
// @Router /secure/authenticate [POST]
func (s Secure) Authenticate(ctx context.Context, req domain.SecureAuthRequest) (*domain.SecureAuthResponse, error) {
	adminId, err := s.service.Authenticate(ctx, req.Token)
	switch {
	case errors.Is(err, domain.ErrTokenExpired):
		return &domain.SecureAuthResponse{
			Authenticated: false,
			ErrorReason:   domain.ErrTokenExpired.Error(),
			AdminId:       0,
		}, nil
	case errors.Is(err, domain.ErrTokenNotFound):
		return &domain.SecureAuthResponse{
			Authenticated: false,
			ErrorReason:   domain.ErrTokenNotFound.Error(),
			AdminId:       0,
		}, nil
	case err != nil:
		return nil, err
	default:
		return &domain.SecureAuthResponse{
			Authenticated: true,
			ErrorReason:   "",
			AdminId:       adminId,
		}, nil
	}
}

// Authorize
// @Tags secure
// @Summary Метод авторизации для администратора
// @Description Проверяет наличие у администратора необходимого разрешения
// @Accept json
// @Produce json
// @Param body body domain.SecureAuthzRequest true "Тело запроса"
// @Success 200 {object} domain.SecureAuthzResponse
// @Failure 500 {object} domain.GrpcError
// @Router /secure/authorize [POST]
func (s Secure) Authorize(ctx context.Context, req domain.SecureAuthzRequest) (*domain.SecureAuthzResponse, error) {
	ok, err := s.service.Authorize(ctx, req.AdminId, req.Permission)
	if err != nil {
		return nil, apierrors.NewInternalServiceError(err)
	}
	return &domain.SecureAuthzResponse{
		Authorized: ok,
	}, nil
}
