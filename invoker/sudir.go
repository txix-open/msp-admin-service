package invoker

import (
	"encoding/base64"
	"fmt"

	"github.com/integration-system/isp-lib/v2/http"
	"msp-admin-service/conf"
	"msp-admin-service/invoker/sudir"
)

var Sudir = sudirCli{
	cli: http.NewJsonRestClient(),
}

type sudirCli struct {
	cli http.RestClient
}

func (c sudirCli) GetToken(cfg conf.SudirAuth, authCode string) (*sudir.TokenResponse, error) {
	method := fmt.Sprintf("%s/blitz/oauth/te?grant_type=authorization_code&code=%s&redirect_uri=%s",
		cfg.Host, authCode, cfg.RedirectURI,
	)
	basicAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", cfg.ClientId, cfg.ClientSecret)))
	headers := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"Authorization": fmt.Sprintf("Basic %s", basicAuth),
	}
	response := new(sudir.TokenResponse)
	err := c.cli.Invoke(http.POST, method, headers, nil, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c sudirCli) GetUser(host, accessToken string) (*sudir.UserResponse, error) {
	method := fmt.Sprintf("%s/blitz/oauth/me", host)
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}
	response := new(sudir.UserResponse)
	err := c.cli.Invoke(http.GET, method, headers, nil, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
