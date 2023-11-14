package conf

import (
	"reflect"

	"github.com/integration-system/isp-kit/dbx"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/rc/schema"
	"github.com/integration-system/jsonschema"
)

// nolint: gochecknoinits
func init() { //
	schema.CustomGenerators.Register("logLevel", func(field reflect.StructField, t *jsonschema.Type) {
		t.Type = "string"
		t.Enum = []interface{}{"debug", "info", "error", "fatal"}
	})
}

type Remote struct {
	Audit     Audit
	Database  dbx.Config
	ExpireSec int      `valid:"required" schema:"Время жизни токена в секундах,in seconds"`
	UiDesign  UIDesign `schema:"Кастомизация интерфейса"`
	//nolint:lll
	IdleTimeoutMs       int                 `schema:"Время бездействия пользователя,в милисекундах, после указанного времени пользователь будет разлогирован из интерфейса в браузере, по умолчанию отключено"`
	SudirAuth           *SudirAuth          `schema:"СУДИР авторизация"`
	LogLevel            log.Level           `schemaGen:"logLevel" schema:"Уровень логирования"`
	AntiBruteforce      AntiBruteforce      `schema:"Настройки антибрут для admin login"`
	BlockInactiveWorker BlockInactiveWorker `valid:"required" schema:"Блокировка неактивных УЗ"`
	Permissions         []Permission        `schema:"Список разрешений"`
	Ldap                *Ldap               `schema:"Настройки LDAP"`
}

type Audit struct {
	EventSettings []AuditEventSetting
	AuditTTl      AuditTTlSetting
}

type AuditTTlSetting struct {
	TimeToLiveInMin       int `valid:"required"`
	ExpireSyncPeriodInMin int `valid:"required"`
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
	ClientId     string `valid:"required~Required"`
	ClientSecret string `valid:"required~Required"`
	Host         string `valid:"required~Required" schema:"Хост, пример https://sudir.mos.ru"`
	RedirectURI  string `valid:"required~Required"`
}

type AntiBruteforce struct {
	MaxInFlightLoginRequests int `valid:"required~Required" schema:"Количество одновременных запросов /login"`
	DelayLoginRequestInSec   int `valid:"required~Required" schema:"Задержка выполнения /login"`
}

type BlockInactiveWorker struct {
	DaysThreshold        int `valid:"required" schema:"Кол-во дней"`
	RunIntervalInMinutes int `valid:"required" schema:"Интервал запуска,в минутах"`
}

type Permission struct {
	Key  string
	Name string
}

type Ldap struct {
	Address  string `valid:"required" schema:"Адрес LDAP"`
	Username string `valid:"required" schema:"Пользователь сервисной УЗ"`
	Password string `valid:"required" schema:"Пароль сервисной УЗ"`
	BaseDn   string `valid:"required"`
}
