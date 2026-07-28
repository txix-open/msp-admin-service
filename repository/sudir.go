package repository

import (
	"context"
	"fmt"

	"msp-admin-service/entity"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/metrics/http_metrics"
)

//nolint:gosec
const (
	getTokenEndpoint = "/oauth/te"
	getUserEndpoint  = "/oauth/me"
)

type Sudir struct {
	cli *httpcli.Client
}

func NewSudir(cli *httpcli.Client) Sudir {
	return Sudir{
		cli: cli,
	}
}

func (s Sudir) GetToken(ctx context.Context, clientSetting entity.ClientSetting) (*entity.SudirTokenResponse, error) {
	ctx = http_metrics.ClientEndpointToContext(ctx, getTokenEndpoint)

	response := entity.SudirTokenResponse{}
	err := s.cli.Post(getTokenEndpoint).
		BasicAuth(httpcli.BasicAuth{
			Username: clientSetting.ClientId,
			Password: clientSetting.ClientSecret,
		}).
		Header("Content-Type", "application/x-www-form-urlencoded").
		QueryParams(map[string]any{
			"grant_type":   "authorization_code",
			"code":         clientSetting.AuthCode,
			"redirect_uri": clientSetting.RedirectURI,
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
