package service

import (
	"context"

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
	CreateUser(ctx context.Context, user entity.CreateUser) (*entity.User, error)
	UpdateUser(ctx context.Context, id int64, user entity.UpdateUser) (*entity.User, error)
	DeleteUser(ctx context.Context, ids []int64) (int, error)
}

type userRoleRepo interface {
	GetRoleById(ctx context.Context, id int) (*entity.Role, error)
}

type User struct {
	userRepo     userRepo
	userRoleRepo userRoleRepo
	logger       log.Logger
}

func NewUser(userRepo userRepo, userRoleRepo userRoleRepo, logger log.Logger) User {
	return User{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		logger:       logger,
	}
}

func (u User) GetProfileById(ctx context.Context, userId int64) (*domain.AdminUserShort, error) {
	user, err := u.userRepo.GetUserById(ctx, userId)
	if err != nil {
		return nil, errors.WithMessagef(err, "get user by id: %d", userId)
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

func (u User) GetUsers(ctx context.Context, identities domain.UsersRequest) (*domain.UsersResponse, error) {
	users, err := u.userRepo.GetUsers(ctx, identities.Ids, identities.Offset, identities.Limit, identities.Email)
	if err != nil {
		return nil, errors.WithMessage(err, "get users from repo")
	}

	items := make([]domain.User, 0, len(users))
	for _, user := range users {
		items = append(items, domain.User(user))
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

	user, err = u.userRepo.CreateUser(ctx, entity.CreateUser{
		RoleId:    req.RoleId,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  encryptedPassword,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "create user")
	}

	result := domain.User(*user)

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
			return nil, errors.WithMessage(err, "")
		}
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

	result := domain.User(*user)

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

//nolint:gomnd
func (u User) cryptPassword(password string) (string, error) {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", errors.WithMessage(err, "gen bcrypt from password")
	}

	return string(passwordBytes), nil
}
