package conf

import (
	"reflect"

	"github.com/integration-system/isp-kit/dbx"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/rc/schema"
	"github.com/integration-system/jsonschema"
)

func init() {
	schema.CustomGenerators.Register("logLevel", func(field reflect.StructField, t *jsonschema.Type) {
		t.Type = "string"
		t.Enum = []interface{}{"debug", "info", "error", "fatal"}
	})
}

type Remote struct {
	Database  dbx.Config
	ExpireSec int        `valid:"required" schema:"Время жизни токена в секундах,in seconds"`
	UiDesign  UIDesign   `schema:"Кастомизация интерфейса"`
	SudirAuth *SudirAuth `schema:"СУДИР авторизация"`
	LogLevel  log.Level  `schemaGen:"logLevel" schema:"Уровень логирования"`
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
