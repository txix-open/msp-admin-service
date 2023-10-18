package ldap

import (
	"context"
	"fmt"

	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
	"msp-admin-service/conf"
	"msp-admin-service/entity"
)

type Repo interface {
	IsExist(ctx context.Context, dn string) (bool, error)
	DnByUserPrincipalName(ctx context.Context, principalName string) (string, error)
	RemoveFromGroup(ctx context.Context, userDn string, groupDn string) error
	Close() error
}

type RoleRepo interface {
	All(ctx context.Context) ([]entity.Role, error)
}

type UserRoleRepo interface {
	GetRolesByUserIds(ctx context.Context, identity []int) ([]entity.UserRole, error)
	UpdateUserRoleLinks(ctx context.Context, id int, roleIds []int) error
}

type RepoSupplier func(config *conf.Ldap) (Repo, error)

type Service struct {
	config       *conf.Ldap
	repoSuppler  RepoSupplier
	userRoleRepo UserRoleRepo
	roleRepo     RoleRepo
	logger       log.Logger
}

func NewService(
	config *conf.Ldap,
	repoSuppler RepoSupplier,
	userRoleRepo UserRoleRepo,
	roleRepo RoleRepo,
	logger log.Logger,
) Service {
	return Service{
		config:       config,
		repoSuppler:  repoSuppler,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
		logger:       logger,
	}
}

func (s Service) RemoveGroups(ctx context.Context, user entity.User) error {
	ctx = log.ToContext(ctx, log.String("process", "ldap"), log.Int64("userId", user.Id))

	ldapRepo, err := s.repoSuppler(s.config)
	if err != nil {
		return errors.WithMessage(err, "init ldap repository")
	}
	defer ldapRepo.Close()

	if user.SudirUserId == nil || *user.SudirUserId == "" {
		s.logger.Info(ctx, "user has not sudir user id, skip")
		return nil
	}

	userDn, err := ldapRepo.DnByUserPrincipalName(ctx, *user.SudirUserId)
	if err != nil {
		return errors.WithMessagef(err, "get user dn by principal name %s", *user.SudirUserId)
	}

	userRoles, err := s.userRoleRepo.GetRolesByUserIds(ctx, []int{int(user.Id)})
	if err != nil {
		return errors.WithMessage(err, "ger roles by user ids")
	}

	allRoles, err := s.roleRepo.All(ctx)
	if err != nil {
		return errors.WithMessage(err, "get all roles")
	}
	rolesById := make(map[int]entity.Role)
	for _, role := range allRoles {
		rolesById[role.Id] = role
	}

	newUserRoles := make([]int, 0)
	for _, role := range userRoles {
		keepRole, err := s.handleRole(ctx, role.RoleId, rolesById, userDn, ldapRepo)
		if err != nil {
			s.logger.Error(ctx, errors.WithMessage(err, "handle role"))
			continue
		}
		if keepRole {
			newUserRoles = append(newUserRoles, role.RoleId)
		}
	}

	err = s.userRoleRepo.UpdateUserRoleLinks(ctx, int(user.Id), newUserRoles)
	if err != nil {
		return errors.WithMessage(err, "update user role links")
	}

	return nil
}

func (s Service) handleRole(
	ctx context.Context,
	roleId int,
	rolesById map[int]entity.Role,
	userDn string,
	ldapRepo Repo,
) (bool, error) {
	role, ok := rolesById[roleId]
	if !ok {
		return false, errors.Errorf("role with id %d not found", roleId)
	}

	if role.ExternalGroup == "" {
		s.logger.Info(ctx, "role has not external group mapping, skip", log.Int("roleId", roleId))
		return true, nil
	}

	exist, err := ldapRepo.IsExist(ctx, role.ExternalGroup)
	if err != nil {
		return false, errors.WithMessagef(err, "check dn %s is exist", role.ExternalGroup)
	}
	if !exist {
		s.logger.Info(ctx, "role doesn't exist in ldap, skip", log.Int("roleId", roleId))
		return true, nil
	}

	err = ldapRepo.RemoveFromGroup(ctx, userDn, role.ExternalGroup)
	if err != nil {
		return false, errors.WithMessagef(err, "remove %s from group %s", userDn, role.ExternalGroup)
	}

	s.logger.Info(ctx, fmt.Sprintf("user %s removed from group %s", userDn, role.ExternalGroup))

	return false, nil
}
