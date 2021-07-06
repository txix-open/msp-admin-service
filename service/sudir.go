package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"msp-admin-service/conf"
	"msp-admin-service/entity"
)

type sudirTokenResponse struct {
	*SudirAuthError

	IdToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

type SudirAuthError struct {
	ErrorName        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (s *SudirAuthError) Error() string {
	return fmt.Sprintf("error: %s; description: %s", s.ErrorName, s.ErrorDescription)
}

type sudirUser struct {
	*SudirAuthError

	UserId     string `json:"uid"`
	Logonname  string `json:"logonname"`
	Firstname  string `json:"firstname"`
	Surname    string `json:"surname"`
	Middlename string `json:"middlename"`
	Email      string `json:"email"`
}

func AuthSudir(cfg conf.SudirAuth, authCode string) (entity.AdminUser, error) {
	tokenResponse, err := getSudirTokens(cfg, authCode)
	if err != nil {
		return entity.AdminUser{}, err
	} else if tokenResponse.SudirAuthError != nil {
		return entity.AdminUser{}, tokenResponse.SudirAuthError
	}

	user, err := getSudirUser(cfg.Host, tokenResponse.AccessToken)
	if err != nil {
		return entity.AdminUser{}, err
	} else if user.SudirAuthError != nil {
		return entity.AdminUser{}, user.SudirAuthError
	}

	return entity.AdminUser{
		SudirUserId: user.UserId,
		FirstName:   user.Firstname,
		LastName:    user.Surname,
		Email:       user.Email,
		Password:    "",
	}, nil
}

func getSudirTokens(cfg conf.SudirAuth, authCode string) (sudirTokenResponse, error) {
	method := cfg.Host + "/blitz/oauth/te"
	payload := url.Values{
		"grant_type":   []string{"authorization_code"},
		"code":         []string{authCode},
		"redirect_uri": []string{cfg.RedirectURI},
	}

	req, err := http.NewRequest("POST", method, strings.NewReader(payload.Encode()))
	if err != nil {
		return sudirTokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(cfg.ClientId, cfg.ClientSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return sudirTokenResponse{}, err
	}
	defer resp.Body.Close()
	var tokenResponse sudirTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		return sudirTokenResponse{}, err
	}

	return tokenResponse, nil
}

func getSudirUser(host, accessToken string) (sudirUser, error) {
	method := host + "/blitz/oauth/me"

	req, err := http.NewRequest("GET", method, nil)
	if err != nil {
		return sudirUser{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return sudirUser{}, err
	}
	defer resp.Body.Close()
	var user sudirUser
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return sudirUser{}, err
	}

	return user, nil
}
