package service

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
)

type auditService interface {
	SaveAuditAsync(ctx context.Context, userId int64, message string)
}

type userRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	UpsertBySudirUserId(ctx context.Context, user entity.User) (*entity.User, error)
}

type tokenService interface {
	GenerateToken(ctx context.Context, id int64) (string, string, error)
	RevokeAllByUserId(ctx context.Context, userId int64) error
}

type sudirService interface {
	Authenticate(ctx context.Context, authCode string) (*entity.SudirUser, error)
}

type Auth struct {
	userRepository           userRepository
	tokenService             tokenService
	sudirService             sudirService
	auditService             auditService
	logger                   log.Logger
	maxInFlightLoginRequests int
	delayLoginRequest        time.Duration
	inFlight                 *atomic.Int32
}

func NewAuth(
	userRepository userRepository,
	tokenService tokenService,
	sudirService sudirService,
	auditService auditService,
	logger log.Logger,
	delayLoginRequestInSec int,
	maxInFlightLoginRequests int,
) Auth {
	return Auth{
		userRepository:           userRepository,
		tokenService:             tokenService,
		sudirService:             sudirService,
		auditService:             auditService,
		logger:                   logger,
		delayLoginRequest:        time.Duration(delayLoginRequestInSec) * time.Second,
		maxInFlightLoginRequests: maxInFlightLoginRequests,
		inFlight:                 &atomic.Int32{},
	}
}

func (a Auth) Login(ctx context.Context, request domain.LoginRequest) (*domain.LoginResponse, error) {
	value := a.inFlight.Add(1)
	defer a.inFlight.Add(-1)

	if value > int32(a.maxInFlightLoginRequests) {
		return nil, domain.ErrTooManyLoginRequests
	}
	time.Sleep(a.delayLoginRequest)

	user, err := a.userRepository.GetUserByEmail(ctx, request.Email)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return nil, errors.WithMessage(domain.ErrUnauthenticated, "wrong email")
	case err != nil:
		return nil, errors.WithMessage(err, "get user by email")
	case user.SudirUserId != nil:
		return nil, domain.ErrSudirAuthorization
	}

	if user.Blocked {
		a.auditService.SaveAuditAsync(ctx, user.Id, "Неуспешный вход. Пользователь заблокирован")
		return nil, errors.WithMessagef(domain.ErrUnauthenticated, "user '%d' is blocked", user.Id)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		a.auditService.SaveAuditAsync(ctx, user.Id, "Неуспешный вход. Неверный пароль")
		return nil, errors.WithMessage(domain.ErrUnauthenticated, "wrong password")
	}

	tokenString, expired, err := a.tokenService.GenerateToken(ctx, user.Id)
	if err != nil {
		return nil, errors.WithMessage(err, "generate token")
	}

	a.auditService.SaveAuditAsync(ctx, user.Id, "Успешный вход через форму входа")

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

	user, err := a.userRepository.UpsertBySudirUserId(ctx, entity.User{
		SudirUserId: &sudirUser.SudirUserId,
		RoleId:      sudirUser.RoleId,
		FirstName:   sudirUser.FirstName,
		LastName:    sudirUser.LastName,
		Email:       sudirUser.Email,
		Password:    "",
		Blocked:     false,
		UpdatedAt:   time.Now().UTC(),
		CreatedAt:   time.Now().UTC(),
	})
	if errors.Is(err, domain.ErrNotFound) {
		return nil, errors.Errorf("user with sudir user id = %s is blocked", sudirUser.SudirUserId)
	}
	if err != nil {
		return nil, errors.WithMessage(err, "upsert by sudir user id")
	}

	tokenString, expired, err := a.tokenService.GenerateToken(ctx, user.Id)
	if err != nil {
		return nil, errors.WithMessage(err, "generate token")
	}

	a.auditService.SaveAuditAsync(ctx, user.Id, "Успешный вход через СУДИР")

	return &domain.LoginResponse{
		Token:      tokenString,
		Expired:    expired,
		HeaderName: domain.AdminAuthHeaderName,
	}, nil
}

func (a Auth) Logout(ctx context.Context, adminId int64) error {
	err := a.tokenService.RevokeAllByUserId(ctx, adminId)
	if err != nil {
		return errors.WithMessage(err, "revoke all tokens by user id")
	}

	a.auditService.SaveAuditAsync(ctx, adminId, "Выход")

	return nil
}
