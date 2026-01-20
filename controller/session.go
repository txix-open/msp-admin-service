package controller

import (
	"context"

	"msp-admin-service/domain"
)

type SessionService interface {
	All(ctx context.Context, req domain.SessionPageRequest) (*domain.SessionResponse, error)
	Revoke(ctx context.Context, id int) error
}

type Session struct {
	service SessionService
}

func NewSession(service SessionService) Session {
	return Session{
		service: service,
	}
}

// All
// @Tags session
// @Summary Получение списка сессий
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.SessionPageRequest true "Тело запроса"
// @Success 200 {object} domain.SessionResponse
// @Failure 400 {object} domain.GrpcError "Невалидное тело запроса"
// @Failure 500 {object} domain.GrpcError
// @Router /session/all [POST]
func (c Session) All(ctx context.Context, req domain.SessionPageRequest) (*domain.SessionResponse, error) {
	return c.service.All(ctx, req)
}

// Revoke
// @Tags session
// @Summary Отзыв сессии
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.RevokeRequest true "Тело запроса"
// @Success 200
// @Failure 400 {object} domain.GrpcError "Невалидное тело запроса"
// @Failure 500 {object} domain.GrpcError
// @Router /session/revoke [POST]
func (c Session) Revoke(ctx context.Context, req domain.RevokeRequest) error {
	return c.service.Revoke(ctx, req.Id)
}
