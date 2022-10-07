package service

import (
	"context"

	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
)

type userRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserBySudirUserId(ctx context.Context, id string) (*entity.User, error)
	CreateSudirUser(ctx context.Context, user entity.SudirUser) (*entity.User, error)
}

type tokenService interface {
	GenerateToken(ctx context.Context, id int64) (string, string, error)
}

type sudirService interface {
	Authenticate(ctx context.Context, authCode string) (*entity.SudirUser, error)
}

type Auth struct {
	userRepository userRepository
	tokenService   tokenService
	sudirService   sudirService
	logger         log.Logger
}

func NewAuth(userRepository userRepository, tokenService tokenService, sudirService sudirService, logger log.Logger) Auth {
	return Auth{
		userRepository: userRepository,
		tokenService:   tokenService,
		sudirService:   sudirService,
		logger:         logger,
	}
}

func (a Auth) Login(ctx context.Context, request domain.LoginRequest) (*domain.LoginResponse, error) {
	user, err := a.userRepository.GetUserByEmail(ctx, request.Email)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return nil, errors.WithMessage(domain.ErrUnauthenticated, "wrong email")
	case err != nil:
		return nil, errors.WithMessage(err, "get user by email")
	case user.SudirUserId != nil:
		return nil, domain.ErrSudirAuthorization
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		return nil, errors.WithMessage(domain.ErrUnauthenticated, "wrong password")
	}

	tokenString, expired, err := a.tokenService.GenerateToken(ctx, user.Id)
	if err != nil {
		return nil, errors.WithMessage(err, "generate token")
	}

	return &domain.LoginResponse{
		Token:      tokenString,
		Expired:    expired,
		HeaderName: domain.AdminAuthHeaderName,
	}, nil
}

func (a Auth) LoginWithSudir(ctx context.Context, request domain.LoginSudirRequest) (*domain.LoginResponse, error) {
	sudirUser, err := a.sudirService.Authenticate(ctx, request.AuthCode)

	var authErr *entity.SudirAuthError
	switch {
	case errors.As(err, &authErr):
		a.logger.Error(ctx, "sudir authenticate: error occurred", log.String("error", err.Error()))
		return nil, domain.ErrUnauthenticated
	case err != nil:
		return nil, errors.WithMessage(err, "sudir authenticate")
	case sudirUser.SudirUserId == "" || sudirUser.Email == "":
		a.logger.Error(ctx, "sudir authenticate: missing sudirUserId or email",
			log.String("userId", sudirUser.SudirUserId),
			log.String("email", sudirUser.Email))
		return nil, domain.ErrUnauthenticated
	}

	user, err := a.userRepository.GetUserBySudirUserId(ctx, sudirUser.SudirUserId)

	switch {
	case errors.Is(err, domain.ErrNotFound):
		user, err = a.userRepository.CreateSudirUser(ctx, *sudirUser)
		if err != nil {
			return nil, errors.WithMessage(err, "create sudir user")
		}
	case err != nil:
		return nil, errors.WithMessage(err, "get user")
	}

	tokenString, expired, err := a.tokenService.GenerateToken(ctx, user.Id)
	if err != nil {
		return nil, errors.WithMessage(err, "generate token")
	}

	return &domain.LoginResponse{
		Token:      tokenString,
		Expired:    expired,
		HeaderName: domain.AdminAuthHeaderName,
	}, nil
}
