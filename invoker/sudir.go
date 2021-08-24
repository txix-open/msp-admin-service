package invoker

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"msp-admin-service/conf"
	"msp-admin-service/invoker/sudir"
)

func GetSudirTokens(cfg conf.SudirAuth, authCode string) (sudir.TokenResponse, error) {
	method := cfg.Host + "/blitz/oauth/te"
	payload := url.Values{
		"grant_type":   []string{"authorization_code"},
		"code":         []string{authCode},
		"redirect_uri": []string{cfg.RedirectURI},
	}

	req, err := http.NewRequest("POST", method, strings.NewReader(payload.Encode()))
	if err != nil {
		return sudir.TokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(cfg.ClientId, cfg.ClientSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return sudir.TokenResponse{}, err
	}
	defer resp.Body.Close()
	var tokenResponse sudir.TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		return sudir.TokenResponse{}, err
	}

	return tokenResponse, nil
}

func GetSudirUser(host, accessToken string) (sudir.UserResponse, error) {
	method := host + "/blitz/oauth/me"

	req, err := http.NewRequest("GET", method, nil)
	if err != nil {
		return sudir.UserResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return sudir.UserResponse{}, err
	}
	defer resp.Body.Close()
	var user sudir.UserResponse
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return sudir.UserResponse{}, err
	}

	return user, nil
}
