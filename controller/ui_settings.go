package controller

import (
	"context"
	"msp-admin-service/domain"
)

type UiSettingsService interface {
	SudirSettings(ctx context.Context, req domain.ClientNameRequest) (*domain.UiSudirSetting, error)
}

type UiSettings struct {
	service UiSettingsService
}

func NewUiSettings(service UiSettingsService) UiSettings {
	return UiSettings{
		service: service,
	}
}

// UiSudirSettings
// @Tags ui_settings
// @Summary Получение настроек для ui для авторизации через СУДИР
// @Description Получение настроек для ui для авторизации через СУДИР
// @Accept json
// @Produce json
// @Param body body domain.ClientNameRequest true "Тело запроса"
// @Success 200 {object} domain.UiSudirSetting
// @Failure 401 {object} domain.GrpcError "Неизвестное название клиента"
// @Failure 500 {object} domain.GrpcError
// @Router /ui_settings/sudir [POST]
func (c UiSettings) SudirSettings(ctx context.Context, req domain.ClientNameRequest) (*domain.UiSudirSetting, error) {
	return c.service.SudirSettings(ctx, req)
}
