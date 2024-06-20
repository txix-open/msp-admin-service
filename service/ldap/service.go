package ldap

import (
	"context"
	"fmt"
	"slices"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
	"msp-admin-service/conf"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
)

type Repo interface {
	IsExist(ctx context.Context, dn string) (bool, error)
	DnByUserPrincipalName(ctx context.Context, principalName string) (string, error)
	ModifyMemberAttr(ctx context.Context, userDn string, groupDn string, operation string) error
	Close() error
}

type RoleRepo interface {
	All(ctx context.Context) ([]entity.Role, error)
}

type UserRoleRepo interface {
	GetRolesByUserIds(ctx context.Context, identity []int) ([]entity.UserRole, error)
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

func (s Service) SyncGroupsAsync(ctx context.Context, user entity.User) {
	ctx = context.WithoutCancel(ctx)
	go func() {
		err := s.SyncGroups(ctx, user)
		if err != nil {
			s.logger.Warn(ctx, errors.WithMessage(err, "ldap sync groups"))
		}
	}()
}

func (s Service) SyncGroups(ctx context.Context, user entity.User) error {
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

	for i := range allRoles {
		role := allRoles[i]
		contains := slices.ContainsFunc(userRoles, func(userRole entity.UserRole) bool {
			return userRole.RoleId == role.Id
		})
		operation := entity.GroupOperationDelete
		if contains {
			operation = entity.GroupOperationAdd
		}

		err := s.handleRole(ctx, role, userDn, operation, ldapRepo)
		if errors.Is(err, domain.ErrNoActionRequired) {
			s.logger.Info(ctx, "no action required, role is already synced")
			continue
		}
		if err != nil {
			s.logger.Error(ctx, errors.WithMessage(err, "handle role"))
			continue
		}
	}

	return nil
}

func (s Service) handleRole(
	ctx context.Context,
	role entity.Role,
	userDn string,
	operation string,
	ldapRepo Repo,
) error {
	if role.ExternalGroup == "" {
		s.logger.Info(ctx, "role has no external group mapping, skip", log.Int("roleId", role.Id))
		return nil
	}

	exist, err := ldapRepo.IsExist(ctx, role.ExternalGroup)
	if err != nil {
		return errors.WithMessagef(err, "check dn %s is exist", role.ExternalGroup)
	}
	if !exist {
		s.logger.Info(ctx, "role doesn't exist in ldap, skip", log.Int("roleId", role.Id))
		return nil
	}

	err = ldapRepo.ModifyMemberAttr(ctx, userDn, role.ExternalGroup, operation)
	if err != nil {
		return errors.WithMessage(err, "modify group member attr")
	}

	s.logger.Info(ctx, fmt.Sprintf("group %s synced for user %s, operation %s", role.ExternalGroup, userDn, operation))

	return nil
}
