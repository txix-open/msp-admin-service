package sudir

type UserResponse struct {
	*SudirAuthError

	Email      string   `json:"email"`
	Groups     []string `json:"groups"`
	Sub        string   `json:"sub"`
	GivenName  string   `json:"given_name"`
	FamilyName string   `json:"family_name"`
}

type TokenResponse struct {
	*SudirAuthError

	IdToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}
