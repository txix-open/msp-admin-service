package controller

import (
	"context"

	"msp-admin-service/domain"
)

type AuditService interface {
	All(ctx context.Context, limit int, offset int) (*domain.AuditResponse, error)
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
// @Param body body domain.AuditResponse true "Тело запроса"
// @Success 200 {object} domain.PageRequest
// @Failure 400 {object} domain.GrpcError "Невалидное тело запроса"
// @Failure 500 {object} domain.GrpcError
// @Router /log/all [POST]
func (c Audit) All(ctx context.Context, req domain.PageRequest) (*domain.AuditResponse, error) {
	return c.service.All(ctx, req.Limit, req.Offset)
}
