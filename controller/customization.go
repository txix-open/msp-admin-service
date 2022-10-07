package controller

import (
	"msp-admin-service/conf"
)

type Customization struct {
	uiCfg conf.UIDesign
}

func NewCustomization(uiCfg conf.UIDesign) Customization {
	return Customization{
		uiCfg: uiCfg,
	}
}

// GetUIDesign
// @Tags user
// @Summary Получение внешнего вида
// @Description Получение внешнего вида (палитра и наименование) админ-интерфейса
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Success 200 {object} conf.UIDesign
// @Failure 400 {object} domain.GrpcError "Невалидный токен"
// @Failure 500 {object} domain.GrpcError
// @Router /user/get_design [POST]
func (c Customization) GetUIDesign() (conf.UIDesign, error) {
	return c.uiCfg, nil
}
