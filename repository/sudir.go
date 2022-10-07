package repository

import (
	"context"
	"fmt"
	"net/url"

	"github.com/integration-system/isp-kit/http/httpcli"
	"github.com/pkg/errors"
	"msp-admin-service/conf"
	"msp-admin-service/entity"
)

type Sudir struct {
	httpCli *httpcli.Client
	cfg     *conf.SudirAuth
}

func NewSudir(httpCli *httpcli.Client, cfg *conf.SudirAuth) Sudir {
	return Sudir{
		httpCli: httpCli,
		cfg:     cfg,
	}
}

func (s Sudir) GetToken(ctx context.Context, authCode string) (*entity.SudirTokenResponse, error) {
	urlString, err := url.JoinPath(s.cfg.Host, "/blitz/oauth/te")
	if err != nil {
		return nil, errors.WithMessage(err, "build url")
	}

	response := entity.SudirTokenResponse{}

	res, err := s.httpCli.Post(urlString).
		BasicAuth(httpcli.BasicAuth{
			Username: s.cfg.ClientId,
			Password: s.cfg.ClientSecret,
		}).
		Header("Content-Type", "application/x-www-form-urlencoded").
		QueryParams(map[string]any{
			"grant_type":   "authorization_code",
			"code":         authCode,
			"redirect_uri": s.cfg.RedirectURI,
		}).
		JsonResponseBody(&response).
		Do(ctx)

	if err != nil {
		return nil, errors.WithMessage(err, "http request")
	}

	defer res.Close()

	if !res.IsSuccess() {
		return nil, errors.Errorf("unexpected response: code: %d, status: %s", res.StatusCode(), res.Raw.Status)
	}

	return &response, nil
}

func (s Sudir) GetUser(ctx context.Context, accessToken string) (*entity.SudirUserResponse, error) {
	urlString, err := url.JoinPath(s.cfg.Host, "/blitz/oauth/me")
	if err != nil {
		return nil, errors.WithMessage(err, "build url")
	}

	response := entity.SudirUserResponse{}

	res, err := s.httpCli.
		Get(urlString).
		Header("Authorization", fmt.Sprintf("Bearer %s", accessToken)).
		JsonResponseBody(&response).
		Do(ctx)

	if err != nil {
		return nil, errors.WithMessage(err, "http request")
	}

	defer res.Close()

	if !res.IsSuccess() {
		return nil, errors.Errorf("unexpected response: code: %d, status: %s", res.StatusCode(), res.Raw.Status)
	}

	return &response, nil
}
