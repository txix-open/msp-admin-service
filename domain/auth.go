package domain

const AdminAuthHeaderName = "x-auth-admin"

type Auth struct {
	Token      string
	Expired    string `json:",omitempty"`
	HeaderName string
}

type AuthRequest struct {
	Email    string `valid:"required"`
	Password string ` valid:"required"`
}

type SudirAuthRequest struct {
	AuthCode string `valid:"required"`
}
