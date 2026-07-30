package conf

import (
	"reflect"

	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/rc/schema"
	"github.com/txix-open/jsonschema"
)

// nolint: gochecknoinits
func init() {
	schema.CustomGenerators.Register("logLevel", func(field reflect.StructField, t *jsonschema.Schema) {
		t.Type = "string"
		t.Enum = []any{"debug", "info", "error", "fatal"}
	})
}

//nolint:lll
type Remote struct {
	Audit               Audit               `schema:"Настройки audit"`
	Database            dbx.Config          `schema:"Настройка подключения к БД"`
	ExpireSec           int                 `validate:"required" schema:"Время жизни токена в секундах,in seconds"`
	UiDesign            UIDesign            `schema:"Кастомизация интерфейса"`
	IdleTimeoutMs       int                 `schema:"Время бездействия пользователя,в милисекундах, после указанного времени пользователь будет разлогирован из интерфейса в браузере, по умолчанию отключено"`
	SudirAuth           *SudirAuth          `schema:"СУДИР авторизация"`
	LogLevel            log.Level           `schemaGen:"logLevel" schema:"Уровень логирования"`
	AntiBruteforce      AntiBruteforce      `schema:"Настройки антибрут для admin login"`
	BlockInactiveWorker BlockInactiveWorker `validate:"required" schema:"Блокировка неактивных УЗ"`
	Permissions         []Permission        `schema:"Список разрешений"`
}

type Audit struct {
	EventSettings []AuditEventSetting `schema:"Маппинг событий"`
	AuditTTl      AuditTTlSetting     `schema:"Настройки TTL для audit"`
}

type AuditTTlSetting struct {
	TimeToLiveInMin       int `validate:"required" schema:"TTL события в минутах"`
	ExpireSyncPeriodInMin int `validate:"required" schema:"Интервал запуска воркера"`
}

type AuditEventSetting struct {
	Event string `schema:"Ключ события"`
	Name  string `schema:"Пользовательское название события"`
}

type UIDesign struct {
	Name         string `schema:"Название стенда"`
	PrimaryColor string `schema:"Цвет в HEX,примеры: #ff4d4f #fa8c16 #a0d911 #1890ff #722ed1 #d4b106 #e91e63 #ff5722 #795548 #abb8c3 #525252 #689f38"`
}

type SudirAuth struct {
	Host    string              `validate:"required" schema:"Базовый путь до методов СУДИР, пример https://sudir.mos.ru"`
	Clients []SudirClientConfig `validate:"omitempty,min=1" schema:"Настройка взаимодействия с СУДИР для логина по authCode"`
}

type SudirClientConfig struct {
	ClientName   string          `validate:"required" schema:"Уникальное имя конфигурации клиента"`
	ClientId     string          `validate:"required" schema:"Идентификатор клиента,выданный СУДИР для basicAuth"`
	ClientSecret string          `validate:"required" schema:"Пароль клиента,выданный СУДИР для basicAuth"`
	RedirectURI  string          `validate:"required" schema:"URI для редиректа,передаётся в запросе токена в query"`
	UiSetting    *UiSudirSetting `schema:"Настройки СУДИР для админки"`
}

type UiSudirSetting struct {
	LoginEndpoint  string `validate:"required" schema:"Эндпоинт для входа через СУДИР в адмиинке,может содержать query параметры,client_id и redirect_uri добавляются автоматически"`
	LogoutEndpoint string `validate:"required" schema:"Эндпоинт для выхода через СУДИР в адмиинке"`
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
	Key  string `schema:"Ключ прав"`
	Name string `schema:"Пользовательское название"`
}
