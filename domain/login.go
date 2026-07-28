package domain

const (
	AdminAuthHeaderName = "x-auth-admin"
	AdminAuthIdHeader   = "x-admin-id"
)

type LogoutRequest struct {
	Reason string // Причина выхода
}

type LoginRequest struct {
	Email    string `validate:"required"`  // Электронная почта
	Password string ` validate:"required"` // Пароль
}

type LoginSudirRequest struct {
	AuthCode   string `validate:"required"` // Код для авторизации в СУДИР
	ClientName string `validate:"required"` // Уникальное имя конфигурации клиента
}

// nolint:tagliatelle,godoclint
type LoginResponse struct {
	Token      string // Токен
	Expired    string `json:",omitempty"` // Дата истечения токена
	HeaderName string // Названия заголовка для передачи токена
}
