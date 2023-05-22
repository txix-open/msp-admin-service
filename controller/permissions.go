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
func (u Permissions) GetPermissions(ctx context.Context) []domain.Permission {
	return u.permissionsService.All(ctx)
}
