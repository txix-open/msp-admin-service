package service

import (
	"context"

	"msp-admin-service/conf"
	"msp-admin-service/domain"
	"msp-admin-service/entity"

	"github.com/pkg/errors"
)

type sudirRepo interface {
	GetToken(ctx context.Context, authCode string) (*entity.SudirTokenResponse, error)
	GetUser(ctx context.Context, accessToken string) (*entity.SudirUserResponse, error)
}

type roleRepo interface {
	GetRolesByExternalGroup(ctx context.Context, groups []string) ([]entity.Role, error)
	InsertRole(ctx context.Context, role entity.Role) (*entity.Role, error)
}

type Sudir struct {
	cfg       *conf.SudirAuth
	sudirRepo sudirRepo
}

func NewSudir(cfg *conf.SudirAuth, sudirRepo sudirRepo) Sudir {
	return Sudir{
		cfg:       cfg,
		sudirRepo: sudirRepo,
	}
}

func (s Sudir) Authenticate(ctx context.Context, authCode string, roleRepo roleRepo) (*entity.SudirUser, error) {
	if s.cfg == nil {
		return nil, domain.ErrSudirAuthIsMissed
	}

	tokenResponse, err := s.sudirRepo.GetToken(ctx, authCode)
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
