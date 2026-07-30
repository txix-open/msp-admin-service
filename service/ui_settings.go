package service

import (
	"context"
	"msp-admin-service/conf"
	"msp-admin-service/domain"
	"net/url"

	"github.com/txix-open/isp-kit/errors"
)

type UiSettings struct {
	sudirBaseUrl string
	settings     map[string]conf.SudirClientConfig
}

func NewUiSettings(sudirBaseUrl string, settings []conf.SudirClientConfig) UiSettings {
	settingsMap := make(map[string]conf.SudirClientConfig, len(settings))
	for _, setting := range settings {
		if setting.UiSetting == nil {
			continue
		}
		settingsMap[setting.ClientName] = setting
	}
	return UiSettings{
		sudirBaseUrl: sudirBaseUrl,
		settings:     settingsMap,
	}
}

func (s UiSettings) SudirSettings(ctx context.Context, req domain.ClientNameRequest) (*domain.UiSudirSetting, error) {
	setting, ok := s.settings[req.ClientName]
	if !ok {
		return &domain.UiSudirSetting{}, nil
	}

	loginUrl, err := s.buildUrl(
		setting.UiSetting.LoginEndpoint,
		map[string]string{
			"client_id":    setting.ClientId,
			"redirect_uri": setting.RedirectURI,
		})
	if err != nil {
		return nil, errors.Errorf("build url for login: %w", err)
	}

	logoutUrl, err := s.buildUrl(setting.UiSetting.LogoutEndpoint, nil)
	if err != nil {
		return nil, errors.Errorf("build url for logout: %w", err)
	}
	return &domain.UiSudirSetting{
		LoginUrl:  loginUrl,
		LogoutUrl: logoutUrl,
	}, nil
}

func (s UiSettings) buildUrl(methodPath string, additionalQuery map[string]string) (string, error) {
	method, err := url.Parse(methodPath)
	if err != nil {
		return "", err
	}

	joined, err := url.JoinPath(s.sudirBaseUrl, method.Path)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(joined)
	if err != nil {
		return "", err
	}

	q := method.Query()
	for k, v := range additionalQuery {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}
