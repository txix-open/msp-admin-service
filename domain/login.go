package domain

const (
	AdminAuthHeaderName = "x-auth-admin"
	AdminAuthIdHeader   = "x-admin-id"
)

type LogoutRequest struct {
	Reason string
}

type LoginRequest struct {
	Email    string `validate:"required"`
	Password string ` validate:"required"`
}

type LoginSudirRequest struct {
	AuthCode string `validate:"required"`
}

// nolint:tagliatelle,godoclint
type LoginResponse struct {
	Token      string
	Expired    string `json:",omitempty"`
	HeaderName string
}
