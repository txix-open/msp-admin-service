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
	sudirRolePrefix = "CN="
)

type sudirRepo interface {
	GetToken(ctx context.Context, authCode string) (*entity.SudirTokenResponse, error)
	GetUser(ctx context.Context, accessToken string) (*entity.SudirUserResponse, error)
}

type roleRepo interface {
	GetRoleByName(ctx context.Context, name string) (*entity.Role, error)
	UpsertRoleByName(ctx context.Context, role entity.Role) (int, error)
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

	//nolint
	/*role := getRole(user.Groups)
	if role == "" {
		return nil, errors.New("undefined role")
	}*/

	roleId, err := s.roleRepo.UpsertRoleByName(ctx, entity.Role{
		Name:          user.GivenName,
		Permissions:   []string{},
		ExternalGroup: user.GivenName,
		ChangeMessage: "",
	})
	if err != nil {
		return nil, errors.WithMessage(err, "upsert role")
	}

	email := user.Email
	if email == "" {
		email = user.Sub
	}

	return &entity.SudirUser{
		RoleIds:     []int{roleId},
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
