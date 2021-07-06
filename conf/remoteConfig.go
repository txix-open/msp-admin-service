package conf

import (
	"github.com/integration-system/isp-lib/v2/structure"
)

type RemoteConfig struct {
	Database  structure.DBConfiguration     `valid:"required~Required" json:"database" schema:"Database"`
	Metrics   structure.MetricConfiguration `schema:"Metrics"`
	SecretKey string                        `valid:"required~Required" schema:"JWT secret"`
	ExpireSec int                           `schema:"Token expire time,in seconds"`
	UiDesign  UIDesign                      `schema:"Кастомизация интерфейса"`
	SudirAuth *SudirAuth                    `schema:"СУДИР авторизация"`
}

type UIDesign struct {
	Name         string `schema:"Название стенда"`
	PrimaryColor string `schema:"Цвет в HEX, примеры: #ff4d4f #fa8c16 #a0d911 #1890ff #722ed1 #d4b106 #e91e63 #ff5722 #795548 #abb8c3 #525252 #689f38"`
}

type SudirAuth struct {
	ClientId     string `valid:"required~Required"`
	ClientSecret string `valid:"required~Required"`
	Host         string `valid:"required~Required" schema:"Хост, пример https://sudir.mos.ru"`
	RedirectURI  string `valid:"required~Required"`
}
