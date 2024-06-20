// nolint:tagliatelle
package domain

const (
	AdminAuthHeaderName = "x-auth-admin"
	AdminAuthIdHeader   = "x-admin-id"
)

type LoginRequest struct {
	Email    string `validate:"required"`
	Password string ` validate:"required"`
}

type LoginSudirRequest struct {
	AuthCode string `validate:"required"`
}

type LoginResponse struct {
	Token      string
	Expired    string `json:",omitempty"`
	HeaderName string
}
