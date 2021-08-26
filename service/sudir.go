package service

import (
	"strings"

	"github.com/pkg/errors"
	"msp-admin-service/conf"
	"msp-admin-service/entity"
	"msp-admin-service/invoker"
	"msp-admin-service/model"
)

const (
	innerAdminRole         = "admin"
	innerReadOnlyAdminRole = "read_only_admin"

	sudirRolePrefix        = "CN="
	sudirAdminRole         = "DIT-KKD-Admins"
	sudirReadOnlyAdminRole = "DIT-KKD-Operators"
)

func AuthSudir(cfg conf.SudirAuth, authCode string) (entity.AdminUser, error) {
	tokenResponse, err := invoker.Sudir.GetToken(cfg, authCode)
	if err != nil {
		return entity.AdminUser{}, err
	} else if tokenResponse.SudirAuthError != nil {
		return entity.AdminUser{}, tokenResponse.SudirAuthError
	}

	user, err := invoker.Sudir.GetUser(cfg.Host, tokenResponse.AccessToken)
	if err != nil {
		return entity.AdminUser{}, err
	} else if user.SudirAuthError != nil {
		return entity.AdminUser{}, user.SudirAuthError
	}

	role := getRole(user.Groups)
	if role == "" {
		return entity.AdminUser{}, errors.New("undefined role")
	}

	roleInfo, err := model.RoleRep.GetRoleByName(role)
	if err != nil {
		return entity.AdminUser{}, errors.WithMessage(err, "get role")
	}
	if roleInfo == nil {
		return entity.AdminUser{}, errors.Errorf("get unknown role: %s", role)
	}

	return entity.AdminUser{
		RoleId:      roleInfo.Id,
		SudirUserId: user.Sub,
		FirstName:   user.GivenName,
		LastName:    user.FamilyName,
		Email:       user.Email,
		Password:    "",
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
