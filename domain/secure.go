package domain

type SecureAuthRequest struct {
	Token string // Токен
}

type SecureAuthResponse struct {
	Authenticated bool   // Флаг успешной аутентикации
	ErrorReason   string // Причина ошибки
	AdminId       int64  // Идентификатор администратора
}

type SecureAuthzRequest struct {
	AdminId    int    // Идентификатор администратора
	Permission string // Роль
}

type SecureAuthzResponse struct {
	Authorized bool // Флаг успешноЙ авторизации
}
