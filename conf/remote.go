package conf

import (
	"reflect"

	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/rc/schema"
	"github.com/txix-open/jsonschema"
)

// nolint: gochecknoinits
func init() { //
	schema.CustomGenerators.Register("logLevel", func(field reflect.StructField, t *jsonschema.Schema) {
		t.Type = "string"
		t.Enum = []interface{}{"debug", "info", "error", "fatal"}
	})
}

type Remote struct {
	Audit     Audit
	Database  dbx.Config
	ExpireSec int      `validate:"required" schema:"Время жизни токена в секундах,in seconds"`
	UiDesign  UIDesign `schema:"Кастомизация интерфейса"`
	//nolint:lll
	IdleTimeoutMs       int                 `schema:"Время бездействия пользователя,в милисекундах, после указанного времени пользователь будет разлогирован из интерфейса в браузере, по умолчанию отключено"`
	SudirAuth           *SudirAuth          `schema:"СУДИР авторизация"`
	LogLevel            log.Level           `schemaGen:"logLevel" schema:"Уровень логирования"`
	AntiBruteforce      AntiBruteforce      `schema:"Настройки антибрут для admin login"`
	BlockInactiveWorker BlockInactiveWorker `validate:"required" schema:"Блокировка неактивных УЗ"`
	Permissions         []Permission        `schema:"Список разрешений"`
	Ldap                *Ldap               `schema:"Настройки LDAP"`
}

type Audit struct {
	EventSettings []AuditEventSetting
	AuditTTl      AuditTTlSetting
}

type AuditTTlSetting struct {
	TimeToLiveInMin       int `validate:"required"`
	ExpireSyncPeriodInMin int `validate:"required"`
}

type AuditEventSetting struct {
	Event string
	Name  string
}

type UIDesign struct {
	Name         string `schema:"Название стенда"`
	PrimaryColor string `schema:"Цвет в HEX, примеры: #ff4d4f #fa8c16 #a0d911 #1890ff #722ed1 #d4b106 #e91e63 #ff5722 #795548 #abb8c3 #525252 #689f38"`
}

type SudirAuth struct {
	ClientId     string `validate:"required"`
	ClientSecret string `validate:"required"`
	Host         string `validate:"required" schema:"Хост, пример https://sudir.mos.ru"`
	RedirectURI  string `validate:"required"`
}

type AntiBruteforce struct {
	MaxInFlightLoginRequests int `validate:"required" schema:"Количество одновременных запросов /login"`
	DelayLoginRequestInSec   int `validate:"required" schema:"Задержка выполнения /login"`
}

type BlockInactiveWorker struct {
	DaysThreshold        int `validate:"required" schema:"Кол-во дней"`
	RunIntervalInMinutes int `validate:"required" schema:"Интервал запуска,в минутах"`
}

type Permission struct {
	Key  string
	Name string
}

type Ldap struct {
	Address  string `validate:"required" schema:"Адрес LDAP"`
	Username string `validate:"required" schema:"Пользователь сервисной УЗ"`
	Password string `validate:"required" schema:"Пароль сервисной УЗ"`
	BaseDn   string `validate:"required"`
}
