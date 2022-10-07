package domain

const AdminAuthHeaderName = "x-auth-admin"

type LoginRequest struct {
	Email    string `valid:"required"`
	Password string ` valid:"required"`
}

type LoginSudirRequest struct {
	AuthCode string `valid:"required"`
}

type LoginResponse struct {
	Token      string
	Expired    string `json:",omitempty"`
	HeaderName string
}
