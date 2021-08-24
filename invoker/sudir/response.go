package sudir

type UserResponse struct {
	*SudirAuthError

	UserId     string `json:"uid"`
	Logonname  string `json:"logonname"`
	Firstname  string `json:"firstname"`
	Surname    string `json:"surname"`
	Middlename string `json:"middlename"`
	Email      string `json:"email"`
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
