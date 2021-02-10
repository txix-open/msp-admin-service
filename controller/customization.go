package controller

import (
	"github.com/integration-system/isp-lib/v2/config"
	"google.golang.org/grpc/metadata"
	"msp-admin-service/conf"
)

// @Tags user
// @Summary Получение внешнего вида
// @Description Получение внешнего вида (палитра и наименование) админ-интерфейса
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Success 200 {object} conf.UIDesign
// @Failure 400 {object} structure.GrpcError "Невалидный токен"
// @Failure 500 {object} structure.GrpcError
// @Router /user/get_design [POST]
func GetUIDesign(_ metadata.MD) (conf.UIDesign, error) {
	return config.GetRemote().(*conf.RemoteConfig).UiDesign, nil
}
