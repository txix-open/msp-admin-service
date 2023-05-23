package service

import (
	"context"
	"fmt"
	"time"

	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
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
}

type TokenRepo interface {
	UpdateStatusByUserId(ctx context.Context, userId int, status string) error
}

type UserRoleRepo interface {
	GetRolesByUserId(ctx context.Context, identity int) ([]int, error)
	GetRolesByUserIds(ctx context.Context, identity []int) ([]entity.UserRole, error)
	InsertPairs(ctx context.Context, id int, roleIds []int) error
	ForceUpsert(ctx context.Context, id int, roleIds []int) error
}

type roleRepoUser interface {
	GetRoleByIds(ctx context.Context, id []int) ([]entity.Role, error)
}

type User struct {
	userRepo     UserRepo
	userRoleRepo UserRoleRepo
	roleRepoUser roleRepoUser
	tokenRepo    TokenRepo
	auditService auditService
	txRunner     UserTransactionRunner
	logger       log.Logger
}

func NewUser(
	userRepo UserRepo,
	userRoleRepo UserRoleRepo,
	roleRepoUser roleRepoUser,
	tokenRepo TokenRepo,
	service auditService,
	txRunner UserTransactionRunner,
	logger log.Logger,
) User {
	return User{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		roleRepoUser: roleRepoUser,
		tokenRepo:    tokenRepo,
		auditService: service,
		txRunner:     txRunner,
		logger:       logger,
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

	roles, err := u.userRoleRepo.GetRolesByUserId(ctx, int(userId))
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return nil, errors.Errorf("unexpected role-user id %d", userId)
	case err != nil:
		return nil, errors.WithMessagef(err, "get role by user id %d", userId)
	}

	var roleList []entity.Role
	if len(roles) != 0 {
		roleList, err = u.roleRepoUser.GetRoleByIds(ctx, roles)
		if err != nil {
			return nil, errors.WithMessage(err, "get roles")
		}
	}

	return &domain.AdminUserShort{
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		Role:        roleList[0].Name,
		Roles:       roles,
		Permissions: mergePermissions(roleList),
	}, nil
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

	items := make([]domain.User, 0, len(users))
	for _, user := range users {
		roles := make([]int, 0)

		for _, role := range rolesByUsers {
			if role.UserId == int(user.Id) {
				roles = append(roles, role.RoleId)
			}
		}

		items = append(items, u.toDomain(user, roles))
	}

	return &domain.UsersResponse{Items: items}, err
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

		if len(req.Roles) != 0 {
			err = tx.InsertPairs(ctx, id, req.Roles)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, errors.WithMessage(err, "create user transaction")
	}

	u.auditService.SaveAuditAsync(ctx, adminId,
		fmt.Sprintf("Пользователь. Создание пользователя %s.", usr.Email),
	)

	result := u.toDomain(usr, req.Roles)
	return &result, nil
}

func (u User) UpdateUser(ctx context.Context, req domain.UpdateUserRequest, adminId int64) (*domain.User, error) {
	var user *entity.User

	err := u.txRunner.UserTransaction(ctx, func(ctx context.Context, tx UserTransaction) error {
		var err error
		user, err = tx.GetUserByEmail(ctx, req.Email)
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return err
		case err != nil:
			return errors.WithMessage(err, "get user by email or phone")
		case user.Id != req.Id:
			return domain.ErrAlreadyExists
		}

		updateEntity := entity.UpdateUser{
			FirstName:            req.FirstName,
			LastName:             req.LastName,
			Email:                req.Email,
			Description:          req.Description,
			LastSessionCreatedAt: req.LastSessionCreatedAt,
		}

		user, err = tx.UpdateUser(ctx, req.Id, updateEntity)
		if err != nil {
			return errors.WithMessage(err, "update user")
		}

		err = tx.ForceUpsert(ctx, int(user.Id), req.Roles)
		if err != nil {
			return errors.WithMessage(err, "force upsert")
		}

		return nil
	})
	if err != nil {
		return nil, errors.WithMessage(err, "update user transaction")
	}

	u.auditService.SaveAuditAsync(ctx, adminId,
		fmt.Sprintf("Пользователь. Изменение пользователя %s.", user.Email),
	)

	result := u.toDomain(*user, req.Roles)
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
		fmt.Sprintf("Пользователь. Удаление пользователей %d.", ids),
	)

	return count, err
}

func (u User) GetById(ctx context.Context, userId int) (*domain.User, error) {
	user, err := u.userRepo.GetUserById(ctx, int64(userId))
	if err != nil {
		return nil, errors.WithMessagef(err, "get user by id %d", userId)
	}

	roles, err := u.userRoleRepo.GetRolesByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	result := u.toDomain(*user, roles)

	return &result, nil
}

func (u User) Block(ctx context.Context, userId int) error {
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

		return nil
	})

	if err != nil {
		return errors.WithMessage(err, "block user transaction")
	}

	return nil
}

//nolint:gomnd
func (u User) cryptPassword(password string) (string, error) {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", errors.WithMessage(err, "gen bcrypt from password")
	}

	return string(passwordBytes), nil
}

func (u User) toDomain(user entity.User, roleIds []int) domain.User {
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
		LastSessionCreatedAt: user.LastSessionCreatedAt,
	}
}
