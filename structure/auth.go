package structure

type Auth struct {
	Token      string `json:"token"`
	Expired    string `json:"expired,omitempty"`
	HeaderName string `json:"headerName"`
}

type AuthRequest struct {
	Email    string `json:"email" valid:"required~Required"`
	Password string `json:"password" valid:"required~Required"`
}
