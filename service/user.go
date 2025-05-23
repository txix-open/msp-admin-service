package service

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
	"golang.org/x/crypto/bcrypt"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
)

type UserTransaction interface {
	UserRepo
	UserRoleRepo
	TokenRepo
}

type UserTransactionRunner interface {
	UserTransaction(ctx context.Context, tx func(ctx context.Context, tx UserTransaction) error) error
}

type UserRepo interface {
	GetUserById(ctx context.Context, identity int64) (*entity.User, error)
	GetUsers(ctx context.Context, ids []int64, offset, limit int, email string) ([]entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	UpdateUser(ctx context.Context, id int64, user entity.UpdateUser) (*entity.User, error)
	DeleteUser(ctx context.Context, ids []int64) (int, error)
	Insert(ctx context.Context, user entity.User) (int, error)
	ChangeBlockStatus(ctx context.Context, userId int) (bool, error)
	ChangePassword(ctx context.Context, userId int64, newPassword string) error
}

type TokenRepo interface {
	UpdateStatusByUserId(ctx context.Context, userId int, status string) error
	LastAccessByUserIds(ctx context.Context, userIds []int) (map[int64]*time.Time, error)
}

type UserRoleRepo interface {
	GetRolesByUserIds(ctx context.Context, identity []int) ([]entity.UserRole, error)
	UpsertUserRoleLinks(ctx context.Context, id int, roleIds []int) error
}

type roleRepoUser interface {
	GetRoleByIds(ctx context.Context, id []int) ([]entity.Role, error)
}

type LdapService interface {
	SyncGroupsAsync(ctx context.Context, user entity.User)
}

type User struct {
	userRepo      UserRepo
	userRoleRepo  UserRoleRepo
	roleRepoUser  roleRepoUser
	tokenRepo     TokenRepo
	auditService  auditService
	txRunner      UserTransactionRunner
	ldapService   LdapService
	tokenService  tokenService
	idleTimeoutMs int
	logger        log.Logger
}

func NewUser(
	userRepo UserRepo,
	userRoleRepo UserRoleRepo,
	roleRepoUser roleRepoUser,
	tokenRepo TokenRepo,
	service auditService,
	txRunner UserTransactionRunner,
	ldapService LdapService,
	tokenService tokenService,
	idleTimeoutMs int,
	logger log.Logger,
) User {
	return User{
		userRepo:      userRepo,
		userRoleRepo:  userRoleRepo,
		roleRepoUser:  roleRepoUser,
		tokenRepo:     tokenRepo,
		auditService:  service,
		txRunner:      txRunner,
		ldapService:   ldapService,
		tokenService:  tokenService,
		idleTimeoutMs: idleTimeoutMs,
		logger:        logger,
	}
}

func (u User) GetProfileById(ctx context.Context, userId int64) (*domain.AdminUserShort, error) {
	user, err := u.userRepo.GetUserById(ctx, userId)
	if err != nil {
		return nil, errors.WithMessagef(err, "get user by id: %d", userId)
	}
	if user.Blocked {
		return nil, errors.WithMessagef(domain.ErrUnauthenticated, "user '%d' is blocked", user.Id)
	}

	roles, err := u.userRoleRepo.GetRolesByUserIds(ctx, []int{int(userId)})
	if err != nil {
		return nil, errors.WithMessagef(err, "get role by user id %d", userId)
	}
	roleIds := RolesIds(roles)

	var (
		roleList []entity.Role
		roleName string
	)
	if len(roleIds) != 0 {
		roleList, err = u.roleRepoUser.GetRoleByIds(ctx, roleIds)
		if err != nil {
			return nil, errors.WithMessage(err, "get roles")
		}
		if len(roleList) > 0 {
			roleName = roleList[0].Name
		}
	}

	return &domain.AdminUserShort{
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Email:         user.Email,
		Role:          roleName,
		Roles:         roleIds,
		IdleTimeoutMs: u.idleTimeoutMs,
		Permissions:   mergePermissions(roleList),
	}, nil
}

func (u User) GetUsers(ctx context.Context, req domain.UsersRequest) (*domain.UsersResponse, error) {
	users, err := u.userRepo.GetUsers(ctx, req.Ids, req.Offset, req.Limit, req.Email)
	if err != nil {
		return nil, errors.WithMessage(err, "get users from repo")
	}

	userIds := make([]int, 0)
	for _, user := range users {
		userIds = append(userIds, int(user.Id))
	}

	rolesByUsers, err := u.userRoleRepo.GetRolesByUserIds(ctx, userIds)
	if err != nil {
		return nil, errors.WithMessage(err, "get roles by user ids")
	}

	lastSessionsByUsers, err := u.tokenRepo.LastAccessByUserIds(ctx, userIds)
	if err != nil {
		return nil, errors.WithMessage(err, "get last sessions by user ids")
	}

	items := make([]domain.User, 0, len(users))
	for _, user := range users {
		roles := make([]int, 0)

		for _, role := range rolesByUsers {
			if role.UserId == int(user.Id) {
				roles = append(roles, role.RoleId)
			}
		}

		items = append(items, u.toDomain(user, roles, lastSessionsByUsers[user.Id]))
	}

	return &domain.UsersResponse{Items: items}, nil
}

func (u User) CreateUser(ctx context.Context, req domain.CreateUserRequest, adminId int64) (*domain.User, error) {
	var usr entity.User

	err := u.txRunner.UserTransaction(ctx, func(ctx context.Context, tx UserTransaction) error {
		user, err := tx.GetUserByEmail(ctx, req.Email)
		switch {
		case errors.Is(err, domain.ErrNotFound):
			break
		case err != nil:
			return errors.WithMessage(err, "get user by email or phone")
		case user != nil:
			return domain.ErrAlreadyExists
		}

		encryptedPassword, err := u.cryptPassword(req.Password)
		if err != nil {
			return errors.WithMessage(err, "crypt password")
		}

		usr = entity.User{
			SudirUserId: nil,
			Id:          0,
			FirstName:   req.FirstName,
			LastName:    req.LastName,
			Email:       req.Email,
			Password:    encryptedPassword,
			Description: req.Description,
			Blocked:     false,
			UpdatedAt:   time.Now().UTC(),
			CreatedAt:   time.Now().UTC(),
		}
		id, err := tx.Insert(ctx, usr)
		if err != nil {
			return errors.WithMessage(err, "create user")
		}

		err = tx.UpsertUserRoleLinks(ctx, id, req.Roles)
		if err != nil {
			return errors.WithMessage(err, "insert user role links")
		}

		return nil
	})
	if err != nil {
		return nil, errors.WithMessage(err, "create user transaction")
	}

	u.ldapService.SyncGroupsAsync(ctx, usr)

	slices.Sort(req.Roles)
	diff := diffToString(map[string]any{
		"Имя":       "",
		"Фамилия":   "",
		"Описание":  "",
		"Email":     "",
		"Роли (ID)": []int{},
	}, map[string]any{
		"Имя":       req.FirstName,
		"Фамилия":   req.LastName,
		"Описание":  req.Description,
		"Email":     req.Email,
		"Роли (ID)": req.Roles,
	})
	u.auditService.SaveAuditAsync(ctx, adminId,
		fmt.Sprintf("Пользователь. Создание пользователя %d. \n %s", usr.Id, diff),
		entity.EventUserChanged,
	)

	result := u.toDomain(usr, req.Roles, nil)
	return &result, nil
}

func (u User) UpdateUser(ctx context.Context, req domain.UpdateUserRequest, adminId int64) (*domain.User, error) {
	var (
		user                 *entity.User
		lastSessionCreatedAt *time.Time
	)
	oldRoles, err := u.userRoleRepo.GetRolesByUserIds(ctx, []int{int(req.Id)})
	if err != nil {
		return nil, errors.WithMessage(err, "get user roles")
	}
	err = u.txRunner.UserTransaction(ctx, func(ctx context.Context, tx UserTransaction) error {
		_, err := tx.GetUserById(ctx, req.Id)
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return err // nolint:wrapcheck
		case err != nil:
			return errors.WithMessage(err, "get user by id")
		}

		user, err = tx.GetUserByEmail(ctx, req.Email)
		switch {
		case errors.Is(err, domain.ErrNotFound):
			break
		case err != nil:
			return errors.WithMessage(err, "get user by email or phone")
		case user.Id != req.Id:
			return domain.ErrAlreadyExists
		}

		updateEntity := entity.UpdateUser{
			FirstName:   req.FirstName,
			LastName:    req.LastName,
			Email:       req.Email,
			Description: req.Description,
		}
		user, err = tx.UpdateUser(ctx, req.Id, updateEntity)
		if err != nil {
			return errors.WithMessage(err, "update user")
		}

		err = tx.UpsertUserRoleLinks(ctx, int(user.Id), req.Roles)
		if err != nil {
			return errors.WithMessage(err, "update user role links")
		}

		userLastSession, err := tx.LastAccessByUserIds(ctx, []int{int(user.Id)})
		if err != nil {
			return errors.WithMessage(err, "get last user session")
		}
		lastSessionCreatedAt = userLastSession[user.Id]

		return nil
	})
	if err != nil {
		return nil, errors.WithMessage(err, "update user transaction")
	}

	u.ldapService.SyncGroupsAsync(ctx, *user)

	slices.Sort(req.Roles)
	diff := diffToString(map[string]any{
		"Имя":       user.FirstName,
		"Фамилия":   user.LastName,
		"Описание":  user.Description,
		"Email":     user.Email,
		"Роли (ID)": RolesIds(oldRoles),
	}, map[string]any{
		"Имя":       req.FirstName,
		"Фамилия":   req.LastName,
		"Описание":  req.Description,
		"Email":     req.Email,
		"Роли (ID)": req.Roles,
	})
	u.auditService.SaveAuditAsync(ctx, adminId,
		fmt.Sprintf("Пользователь. Изменение пользователя %d.\n %s", user.Id, diff),
		entity.EventUserChanged,
	)

	result := u.toDomain(*user, req.Roles, lastSessionCreatedAt)
	return &result, nil
}

func (u User) DeleteUsers(ctx context.Context, ids []int64, adminId int64) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	count, err := u.userRepo.DeleteUser(ctx, ids)
	if err != nil {
		return 0, errors.WithMessage(err, "delete users")
	}

	u.auditService.SaveAuditAsync(ctx, adminId,
		fmt.Sprintf("Пользователь. Удаление пользователей %v.", ids),
		entity.EventUserChanged,
	)

	return count, nil
}

func (u User) GetById(ctx context.Context, userId int) (*domain.User, error) {
	user, err := u.userRepo.GetUserById(ctx, int64(userId))
	if err != nil {
		return nil, errors.WithMessagef(err, "get user by id %d", userId)
	}

	roles, err := u.userRoleRepo.GetRolesByUserIds(ctx, []int{userId})
	if err != nil {
		return nil, errors.WithMessage(err, "get roles by user id")
	}

	lastSessions, err := u.tokenRepo.LastAccessByUserIds(ctx, []int{userId})
	if err != nil {
		return nil, errors.WithMessage(err, "get last sessions by user id")
	}

	result := u.toDomain(*user, RolesIds(roles), lastSessions[int64(userId)])
	return &result, nil
}

func (u User) Block(ctx context.Context, adminId int64, userId int) error {
	userBlocked := "блокировка"

	err := u.txRunner.UserTransaction(ctx, func(ctx context.Context, tx UserTransaction) error {
		blocked, err := tx.ChangeBlockStatus(ctx, userId)
		if err != nil {
			return errors.WithMessage(err, "change block status")
		}

		if blocked {
			err := tx.UpdateStatusByUserId(ctx, userId, entity.TokenStatusRevoked)
			if err != nil {
				return errors.WithMessage(err, "revoke tokens")
			}
		}

		if !blocked {
			userBlocked = "разблокировка"
		}

		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "block user transaction")
	}

	u.auditService.SaveAuditAsync(ctx, adminId,
		fmt.Sprintf("Пользователь. %s пользователя ID %d.", userBlocked, userId),
		entity.EventUserChanged,
	)

	return nil
}

func RolesIds(roles []entity.UserRole) []int {
	roleList := make([]int, 0)

	for _, role := range roles {
		roleList = append(roleList, role.RoleId)
	}
	slices.Sort(roleList)
	return roleList
}

func mergePermissions(roles []entity.Role) []string {
	permissionsMap := make(map[string]bool)
	permList := make([]string, 0)

	for _, role := range roles {
		for _, perm := range role.Permissions {
			if _, ok := permissionsMap[perm]; !ok {
				permissionsMap[perm] = true
				permList = append(permList, perm)
			}
		}
	}

	return permList
}

func diffToString(a map[string]any, b map[string]any) string {
	builder := strings.Builder{}
	for key, aValue := range a {
		bValue := b[key]
		if reflect.DeepEqual(aValue, bValue) {
			continue
		}
		if reflect.ValueOf(bValue).IsZero() {
			builder.WriteString(fmt.Sprintf("%s: %v -> null\n", key, aValue))
			continue
		}
		builder.WriteString(fmt.Sprintf("%s: %v -> %v\n", key, aValue, bValue))
	}

	if builder.Len() == 0 {
		return "Нет изменений"
	}

	result := builder.String()
	return result[:len(result)-1]
}

func (u User) ChangePassword(ctx context.Context, adminId int64, oldPassword string, newPassword string) error {
	admin, err := u.userRepo.GetUserById(ctx, adminId)
	if err != nil {
		return errors.WithMessage(err, "user.service.ChangePassword: get user by id")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(oldPassword)); err != nil {
		u.auditService.SaveAuditAsync(ctx, adminId, "Указан неверный старый пароль", entity.EventErrorPasswordChange)
		return domain.ErrInvalidPassword
	}

	encryptedPassword, err := u.cryptPassword(newPassword)
	if err != nil {
		return errors.WithMessage(err, "user.service.ChangePassword: crypt new password")
	}

	err = u.userRepo.ChangePassword(ctx, adminId, encryptedPassword)
	if err != nil {
		return errors.WithMessage(err, "user.service.ChangePassword: change password ")
	}

	err = u.tokenService.RevokeAllByUserId(ctx, adminId)
	if err != nil {
		return errors.WithMessage(err, "revoke all tokens by user id")
	}

	u.auditService.SaveAuditAsync(ctx, adminId, "Сменил пароль", entity.EventUserPasswordChanged)

	return nil
}

//nolint:gomnd
func (u User) cryptPassword(password string) (string, error) {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12) //nolint:mnd
	if err != nil {
		return "", errors.WithMessage(err, "gen bcrypt from password")
	}

	return string(passwordBytes), nil
}

func (u User) toDomain(user entity.User, roleIds []int, lastSessionCreatedAt *time.Time) domain.User {
	return domain.User{
		Id:                   user.Id,
		Roles:                roleIds,
		FirstName:            user.FirstName,
		Description:          user.Description,
		LastName:             user.LastName,
		Email:                user.Email,
		Blocked:              user.Blocked,
		UpdatedAt:            user.UpdatedAt,
		CreatedAt:            user.CreatedAt,
		LastSessionCreatedAt: lastSessionCreatedAt,
	}
}
