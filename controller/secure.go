package controller

import (
	"context"

	"github.com/pkg/errors"
	"msp-admin-service/domain"
)

type SecureService interface {
	GetUserId(ctx context.Context, token string) (int64, error)
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
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.SecureAuthRequest true "Тело запроса"
// @Success 200 {object} domain.SecureAuthResponse
// @Failure 500 {object} domain.GrpcError
// @Router /secure/authenticate [POST]
func (s Secure) Authenticate(ctx context.Context, req domain.SecureAuthRequest) (*domain.SecureAuthResponse, error) {
	adminId, err := s.service.GetUserId(ctx, req.Token)
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
