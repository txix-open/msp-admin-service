// nolint:tagliatelle
package entity

import (
	"fmt"
)

type SudirAuthError struct {
	ErrorName        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (s *SudirAuthError) Error() string {
	return fmt.Sprintf("error: %s; description: %s", s.ErrorName, s.ErrorDescription)
}

type SudirUserResponse struct {
	*SudirAuthError

	Email      string   `json:"email"`
	Groups     []string `json:"groups"`
	Sub        string   `json:"sub"`
	GivenName  string   `json:"given_name"`
	FamilyName string   `json:"family_name"`
}

type SudirTokenResponse struct {
	*SudirAuthError

	IdToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}
