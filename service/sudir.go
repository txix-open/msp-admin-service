package service

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"msp-admin-service/conf"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
)

const (
	innerAdminRole         = "admin"
	innerReadOnlyAdminRole = "read_only_admin"

	sudirRolePrefix        = "CN="
	sudirAdminRole         = "DIT-KKD-Admins"
	sudirReadOnlyAdminRole = "DIT-KKD-Operators"
)

type sudirRepo interface {
	GetToken(ctx context.Context, authCode string) (*entity.SudirTokenResponse, error)
	GetUser(ctx context.Context, accessToken string) (*entity.SudirUserResponse, error)
}

type roleRepo interface {
	GetRoleByName(ctx context.Context, name string) (*entity.Role, error)
}

type Sudir struct {
	cfg       *conf.SudirAuth
	sudirRepo sudirRepo
	roleRepo  roleRepo
}

func NewSudir(cfg *conf.SudirAuth, sudirRepo sudirRepo, roleRepo roleRepo) Sudir {
	return Sudir{
		cfg:       cfg,
		sudirRepo: sudirRepo,
		roleRepo:  roleRepo,
	}
}

func (s Sudir) Authenticate(ctx context.Context, authCode string) (*entity.SudirUser, error) {
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

	role := getRole(user.Groups)
	if role == "" {
		return nil, errors.New("undefined role")
	}

	roleInfo, err := s.roleRepo.GetRoleByName(ctx, role)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return nil, errors.WithMessagef(err, "get unknown role '%s'", role)
	case err != nil:
		return nil, errors.WithMessage(err, "get role")
	}

	return &entity.SudirUser{
		RoleId:      roleInfo.Id,
		SudirUserId: user.Sub,
		FirstName:   user.GivenName,
		LastName:    user.FamilyName,
		Email:       user.Email,
	}, nil
}

func getRole(groups []string) string {
	var role string
	for _, group := range groups {
		part := strings.Split(group, ",")
		for _, p := range part {
			sudirRole := strings.TrimPrefix(p, sudirRolePrefix)
			switch sudirRole {
			case sudirAdminRole:
				return innerAdminRole
			case sudirReadOnlyAdminRole:
				role = innerReadOnlyAdminRole
			}
		}
	}
	return role
}
