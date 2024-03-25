package domain

type SecureAuthRequest struct {
	Token string
}

type SecureAuthResponse struct {
	Authenticated bool
	ErrorReason   string
	AdminId       int64
}

type SecureAuthzRequest struct {
	AdminId    int
	Permission string
}

type SecureAuthzResponse struct {
	Authorized bool
}
