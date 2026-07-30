package domain

type ClientNameRequest struct {
	ClientName string `validate:"required"`
}

type UiSudirSetting struct {
	LoginUrl  string // URL для получения authCode для авторизации через СУДИР
	LogoutUrl string // URL для завершения сессии
}
