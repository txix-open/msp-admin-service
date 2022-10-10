package domain

type SecureAuthRequest struct {
	Token string
}

type SecureAuthResponse struct {
	Authenticated bool
	ErrorReason   string
	AdminId       int64
}
