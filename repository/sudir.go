package repository

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/metrics/http_metrics"
	"msp-admin-service/conf"
	"msp-admin-service/entity"
)

const (
	getTokenEndpoint = "/oauth/te"
	getUserEndpoint  = "/oauth/me"
)

type Sudir struct {
	cli *httpcli.Client
	cfg *conf.SudirAuth
}

func NewSudir(httpCli *httpcli.Client, cfg *conf.SudirAuth) Sudir {
	return Sudir{
		cli: httpCli,
		cfg: cfg,
	}
}

func (s Sudir) GetToken(ctx context.Context, authCode string) (*entity.SudirTokenResponse, error) {
	ctx = http_metrics.ClientEndpointToContext(ctx, getTokenEndpoint)

	response := entity.SudirTokenResponse{}
	err := s.cli.Post(getTokenEndpoint).
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
		StatusCodeToError().
		DoWithoutResponse(ctx)

	if err != nil {
		return nil, errors.WithMessagef(err, "call POST: '%s'", getTokenEndpoint)
	}

	return &response, nil
}

func (s Sudir) GetUser(ctx context.Context, accessToken string) (*entity.SudirUserResponse, error) {
	ctx = http_metrics.ClientEndpointToContext(ctx, getUserEndpoint)

	response := entity.SudirUserResponse{}
	err := s.cli.Get(getUserEndpoint).
		Header("Authorization", fmt.Sprintf("Bearer %s", accessToken)).
		JsonResponseBody(&response).
		StatusCodeToError().
		DoWithoutResponse(ctx)
	if err != nil {
		return nil, errors.WithMessagef(err, "call GET: '%s'", getUserEndpoint)
	}

	return &response, nil
}
