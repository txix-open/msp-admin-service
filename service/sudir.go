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
	sudirRolePrefix = "CN=" // nolint:unused
)

type sudirRepo interface {
	GetToken(ctx context.Context, authCode string) (*entity.SudirTokenResponse, error)
	GetUser(ctx context.Context, accessToken string) (*entity.SudirUserResponse, error)
}

type roleRepo interface {
	GetRoleByExternalGroup(ctx context.Context, group string) (*entity.Role, error)
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

	var role *entity.Role
	role, err = roleRepo.GetRoleByExternalGroup(ctx, user.GivenName)
	if errors.Is(err, domain.ErrNotFound) {
		role, err = roleRepo.InsertRole(ctx, entity.Role{
			Name:          user.GivenName,
			ExternalGroup: user.GivenName, // TODO take group name correctly from user.Group
			Permissions:   []string{},
		})
		if err != nil {
			return nil, errors.WithMessage(err, "insert role")
		}
	}
	if err != nil {
		return nil, errors.WithMessage(err, "get role by external group")
	}

	return &entity.SudirUser{
		RoleIds:     []int{role.Id},
		SudirUserId: user.Sub,
		FirstName:   user.GivenName,
		LastName:    user.FamilyName,
		Email:       email,
	}, nil
}

// nolint
func getRole(groups []string) string {
	for _, group := range groups {
		part := strings.Split(group, ",")
		for _, p := range part {
			sudirRole := strings.TrimPrefix(p, sudirRolePrefix)
			return sudirRole
		}
	}
	return ""
}
