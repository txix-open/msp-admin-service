package controller

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"msp-admin-service/domain"
)

type AuditService interface {
	All(ctx context.Context, limit int, offset int) (*domain.AuditResponse, error)
	Events(ctx context.Context) ([]domain.AuditEvent, error)
	SetEvents(ctx context.Context, req []domain.SetAuditEvent) error
}

type Audit struct {
	service AuditService
}

func NewAudit(service AuditService) Audit {
	return Audit{
		service: service,
	}
}

// All
// @Tags log
// @Summary Получение списка логов
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.PageRequest true "Тело запроса"
// @Success 200 {object} domain.AuditResponse
// @Failure 400 {object} domain.GrpcError "Невалидное тело запроса"
// @Failure 500 {object} domain.GrpcError
// @Router /log/all [POST]
func (c Audit) All(ctx context.Context, req domain.PageRequest) (*domain.AuditResponse, error) {
	return c.service.All(ctx, req.Limit, req.Offset)
}

// Events
// @Tags log
// @Summary Получение списка логов
// @Description Возвращает полный список доступных событий аудита
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.AuditResponse true "Тело запроса"
// @Success 200 {array} domain.AuditEvent
// @Failure 400 {object} domain.GrpcError "Невалидное тело запроса"
// @Failure 500 {object} domain.GrpcError
// @Router /log/events [POST]
func (c Audit) Events(ctx context.Context) ([]domain.AuditEvent, error) {
	return c.service.Events(ctx)
}

// SetEvents
// @Tags log
// @Summary Получение списка логов
// @Description Всегда возвращает полный список доступных событий аудита
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.SetAuditEvent true "Тело запроса"
// @Success 200
// @Failure 400 {object} domain.GrpcError "Невалидное тело запроса"
// @Failure 500 {object} domain.GrpcError
// @Router /log/set_events [POST]
func (c Audit) SetEvents(ctx context.Context, req []domain.SetAuditEvent) error {
	err := c.service.SetEvents(ctx, req)
	switch {
	case errors.As(err, &domain.UnknownAuditEventError{}):
		return status.Errorf(codes.InvalidArgument, err.Error())
	default:
		return err
	}
}
