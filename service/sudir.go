package service

import (
	"context"

	"msp-admin-service/conf"
	"msp-admin-service/domain"
	"msp-admin-service/entity"

	"github.com/pkg/errors"
)

type sudirRepo interface {
	GetToken(ctx context.Context, creds entity.ClientSetting) (*entity.SudirTokenResponse, error)
	GetUser(ctx context.Context, accessToken string) (*entity.SudirUserResponse, error)
}

type roleRepo interface {
	GetRolesByExternalGroup(ctx context.Context, groups []string) ([]entity.Role, error)
	InsertRole(ctx context.Context, role entity.Role) (*entity.Role, error)
}

type Sudir struct {
	sudirRepo      sudirRepo
	clientSettings map[string]conf.SudirClientConfig
}

func NewSudir(cfg *conf.SudirAuth, sudirRepo sudirRepo) (Sudir, error) {
	settings := make([]conf.SudirClientConfig, 0)
	if cfg != nil {
		settings = cfg.Clients
	}

	clientSettings := make(map[string]conf.SudirClientConfig, len(settings))
	for i, setting := range settings {
		_, ok := clientSettings[setting.ClientName]
		if ok {
			return Sudir{}, errors.Errorf("client setting [%d] has non-unique client name", i)
		}
		clientSettings[setting.ClientName] = setting
	}

	return Sudir{
		sudirRepo:      sudirRepo,
		clientSettings: clientSettings,
	}, nil
}

func (s Sudir) Authenticate(ctx context.Context, req domain.LoginSudirRequest, roleRepo roleRepo) (*entity.SudirUser, error) {
	if len(s.clientSettings) == 0 {
		return nil, domain.ErrSudirAuthIsMissed
	}

	setting, ok := s.clientSettings[req.ClientName]
	if !ok {
		return nil, domain.ErrUnauthenticated
	}

	creds := entity.ClientSetting{
		AuthCode:     req.AuthCode,
		ClientId:     setting.ClientId,
		ClientSecret: setting.ClientSecret,
		RedirectURI:  setting.RedirectURI,
	}
	tokenResponse, err := s.sudirRepo.GetToken(ctx, creds)
	switch {
	case err != nil:
		return nil, errors.WithMessage(err, "get token")
	case tokenResponse.SudirAuthError != nil:
		return nil, errors.WithMessage(tokenResponse.SudirAuthError, "get token")
	}

	user, err := s.sudirRepo.GetUser(ctx, tokenResponse.AccessToken)
	switch {
	case err != nil:
		return nil, errors.WithMessage(err, "get user")
	case user.SudirAuthError != nil:
		return nil, errors.WithMessage(user.SudirAuthError, "get user")
	}
	email := user.Email
	if email == "" {
		email = user.Sub
	}

	rolesIds := make([]int, 0)
	if len(user.Groups) > 0 {
		roles, err := roleRepo.GetRolesByExternalGroup(ctx, user.Groups)
		if err != nil {
			return nil, errors.WithMessage(err, "get roles by external groups")
		}
		for _, role := range roles {
			rolesIds = append(rolesIds, role.Id)
		}
	}

	return &entity.SudirUser{
		RoleIds:     rolesIds,
		SudirUserId: user.Sub,
		FirstName:   user.GivenName,
		LastName:    user.FamilyName,
		FullName:    user.Name,
		Email:       email,
	}, nil
}
