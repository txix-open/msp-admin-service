package controller

import (
	"context"

	"msp-admin-service/domain"
)

type permissionsService interface {
	All(ctx context.Context) []domain.Permission
}

type Permissions struct {
	permissionsService permissionsService
}

func NewPermissions(permissionsService permissionsService) Permissions {
	return Permissions{
		permissionsService: permissionsService,
	}
}

// GetPermissions
// @Tags user
// @Summary Получить все разрашения
// @Description Получить все разрашения
// @Accept json
// @Produce json
// @Success 200 {array} domain.Permission
// @Router /user/get_permissions [POST]
func (u Permissions) GetPermissions(ctx context.Context) []domain.Permission {
	return u.permissionsService.All(ctx)
}
