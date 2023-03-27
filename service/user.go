package service

import (
	"context"
	"time"

	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
)

type userRepo interface {
	GetUserById(ctx context.Context, identity int64) (*entity.User, error)
	GetUsers(ctx context.Context, ids []int64, offset, limit int, email string) ([]entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	UpdateUser(ctx context.Context, id int64, user entity.UpdateUser) (*entity.User, error)
	DeleteUser(ctx context.Context, ids []int64) (int, error)
	Insert(ctx context.Context, user entity.User) (int, error)
	ChangeBlockStatus(ctx context.Context, userId int) (bool, error)
}

type tokenRepo interface {
	UpdateStatusByUserId(ctx context.Context, userId int, status string) error
}

type userRoleRepo interface {
	GetRoleById(ctx context.Context, id int) (*entity.Role, error)
	All(ctx context.Context) ([]entity.Role, error)
}

type User struct {
	userRepo     userRepo
	userRoleRepo userRoleRepo
	tokenRepo    tokenRepo
	logger       log.Logger
}

func NewUser(userRepo userRepo, userRoleRepo userRoleRepo, tokenRepo tokenRepo, logger log.Logger) User {
	return User{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		tokenRepo:    tokenRepo,
		logger:       logger,
	}
}

func (u User) GetProfileById(ctx context.Context, userId int64) (*domain.AdminUserShort, error) {
	user, err := u.userRepo.GetUserById(ctx, userId)
	if err != nil {
		return nil, errors.WithMessagef(err, "get user by id: %d", userId)
	}
	if user.Blocked {
		return nil, errors.Errorf("user is blocked")
	}

	role, err := u.userRoleRepo.GetRoleById(ctx, user.RoleId)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return nil, errors.Errorf("unexpected role id %d", user.RoleId)
	case err != nil:
		return nil, errors.WithMessagef(err, "get role by id %d", user.RoleId)
	}

	return &domain.AdminUserShort{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      role.Name,
	}, nil
}

func (u User) GetUsers(ctx context.Context, req domain.UsersRequest) (*domain.UsersResponse, error) {
	users, err := u.userRepo.GetUsers(ctx, req.Ids, req.Offset, req.Limit, req.Email)
	if err != nil {
		return nil, errors.WithMessage(err, "get users from repo")
	}

	roleNames, err := u.roleNames(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "role names")
	}

	items := make([]domain.User, 0, len(users))
	for _, user := range users {
		items = append(items, u.toDomain(user, roleNames))
	}

	return &domain.UsersResponse{Items: items}, err
}

func (u User) CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error) {
	user, err := u.userRepo.GetUserByEmail(ctx, req.Email)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		break
	case err != nil:
		return nil, errors.WithMessage(err, "get user by email or phone")
	case user != nil:
		return nil, domain.ErrAlreadyExists
	}

	encryptedPassword, err := u.cryptPassword(req.Password)
	if err != nil {
		return nil, errors.WithMessage(err, "crypt password")
	}

	roleNames, err := u.roleNames(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "role names")
	}

	usr := entity.User{
		SudirUserId: nil,
		Id:          0,
		RoleId:      req.RoleId,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		Password:    encryptedPassword,
		Blocked:     false,
		UpdatedAt:   time.Now().UTC(),
		CreatedAt:   time.Now().UTC(),
	}
	id, err := u.userRepo.Insert(ctx, usr)
	if err != nil {
		return nil, errors.WithMessage(err, "create user")
	}
	usr.Id = int64(id)
	result := u.toDomain(usr, roleNames)

	return &result, nil
}

func (u User) UpdateUser(ctx context.Context, req domain.UpdateUserRequest) (*domain.User, error) {
	user, err := u.userRepo.GetUserById(ctx, req.Id)
	switch {
	case err != nil:
		return nil, errors.WithMessage(err, "get user")
	case req.Password != "" && user.SudirUserId != nil:
		return nil, domain.ErrInvalid
	}

	if req.Email != "" && req.Email != user.Email {
		knownUser, err := u.userRepo.GetUserByEmail(ctx, req.Email)
		switch {
		case err != nil:
			return nil, errors.WithMessage(err, "get user by email")
		case knownUser != nil:
			return nil, domain.ErrAlreadyExists
		}
	}

	if req.Password != "" {
		req.Password, err = u.cryptPassword(req.Password)
		if err != nil {
			return nil, errors.WithMessage(err, "encrypt password")
		}
	}

	roleNames, err := u.roleNames(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "role names")
	}

	updateEntity := entity.UpdateUser{
		RoleId:    req.RoleId,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	}

	user, err = u.userRepo.UpdateUser(ctx, req.Id, updateEntity)
	if err != nil {
		return nil, errors.WithMessage(err, "update user")
	}

	result := u.toDomain(*user, roleNames)

	return &result, nil
}

func (u User) DeleteUsers(ctx context.Context, ids []int64) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	count, err := u.userRepo.DeleteUser(ctx, ids)
	if err != nil {
		return 0, errors.WithMessage(err, "delete users")
	}
	return count, err
}

func (u User) GetById(ctx context.Context, userId int) (*domain.User, error) {
	roleNames, err := u.roleNames(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "role names")
	}

	user, err := u.userRepo.GetUserById(ctx, int64(userId))
	if err != nil {
		return nil, errors.WithMessagef(err, "get user by id %d", userId)
	}

	result := u.toDomain(*user, roleNames)

	return &result, nil
}

func (u User) Block(ctx context.Context, userId int) error {
	blocked, err := u.userRepo.ChangeBlockStatus(ctx, userId)
	if err != nil {
		return errors.WithMessage(err, "change block status")
	}

	if blocked {
		err := u.tokenRepo.UpdateStatusByUserId(ctx, userId, entity.TokenStatusRevoked)
		if err != nil {
			return errors.WithMessage(err, "revoke tokens")
		}
	}

	return nil
}

func (u User) Roles(ctx context.Context) ([]domain.Role, error) {
	roles, err := u.userRoleRepo.All(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get all roles")
	}

	result := make([]domain.Role, 0)
	for _, role := range roles {
		desc := ""
		if role.Description != nil {
			desc = *role.Description
		}
		result = append(result, domain.Role{
			Id:          role.Id,
			Name:        role.Name,
			Description: desc,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		})
	}

	return result, nil
}

//nolint:gomnd
func (u User) cryptPassword(password string) (string, error) {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", errors.WithMessage(err, "gen bcrypt from password")
	}

	return string(passwordBytes), nil
}

func (u User) toDomain(user entity.User, roleNames map[int]string) domain.User {
	return domain.User{
		Id:        user.Id,
		RoleId:    user.RoleId,
		RoleName:  roleNames[user.RoleId],
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Blocked:   user.Blocked,
		UpdatedAt: user.UpdatedAt,
		CreatedAt: user.CreatedAt,
	}
}

func (u User) roleNames(ctx context.Context) (map[int]string, error) {
	roles, err := u.userRoleRepo.All(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get all roles")
	}
	result := make(map[int]string)
	for _, role := range roles {
		result[role.Id] = role.Name
	}
	return result, nil
}
