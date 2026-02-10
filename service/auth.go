package service

import (
	"context"
	"sync/atomic"
	"time"

	"msp-admin-service/domain"
	"msp-admin-service/entity"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
	"golang.org/x/crypto/bcrypt"
)

type AuthTransaction interface {
	userRepository
	roleRepo
	UserRoleRepo
	TokenSaver
}

type AuthTransactionRunner interface {
	AuthTransaction(ctx context.Context, tx func(ctx context.Context, tx AuthTransaction) error) error
}

type auditService interface {
	SaveAuditAsync(ctx context.Context, userId int64, message string, event string)
}

type userRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	UpsertBySudirUserId(ctx context.Context, user entity.User) (*entity.User, error)
	UpdateUser(ctx context.Context, id int64, user entity.UpdateUser) (*entity.User, error)
	UpdateLastActiveAt(ctx context.Context, userId int64, lastActiveAt time.Time) error
}

type tokenService interface {
	GenerateToken(ctx context.Context, repo TokenSaver, id int64) (string, string, error)
	RevokeAllByUserId(ctx context.Context, userId int64) error
}

type sudirService interface {
	Authenticate(ctx context.Context, authCode string, repo roleRepo) (*entity.SudirUser, error)
}

type Auth struct {
	userRepository           userRepository
	txRunner                 AuthTransactionRunner
	tokenService             tokenService
	sudirService             sudirService
	auditService             auditService
	logger                   log.Logger
	maxInFlightLoginRequests int64
	delayLoginRequest        time.Duration
	inFlightLoginRequests    *atomic.Int64
}

func NewAuth(
	userRepository userRepository,
	txRunner AuthTransactionRunner,
	tokenService tokenService,
	sudirService sudirService,
	auditService auditService,
	logger log.Logger,
	delayLoginRequestInSec int,
	maxInFlightLoginRequests int,
) Auth {
	return Auth{
		userRepository:           userRepository,
		txRunner:                 txRunner,
		tokenService:             tokenService,
		sudirService:             sudirService,
		auditService:             auditService,
		logger:                   logger,
		delayLoginRequest:        time.Duration(delayLoginRequestInSec) * time.Second,
		maxInFlightLoginRequests: int64(maxInFlightLoginRequests),
		inFlightLoginRequests:    &atomic.Int64{},
	}
}

func (a Auth) Login(ctx context.Context, request domain.LoginRequest) (*domain.LoginResponse, error) {
	value := a.inFlightLoginRequests.Add(1)
	defer a.inFlightLoginRequests.Add(-1)

	if value > a.maxInFlightLoginRequests {
		return nil, domain.ErrTooManyLoginRequests
	}
	time.Sleep(a.delayLoginRequest)

	var (
		tokenString string
		expired     string
	)

	err := a.txRunner.AuthTransaction(ctx, func(ctx context.Context, tx AuthTransaction) error {
		user, err := tx.GetUserByEmail(ctx, request.Email)
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return errors.WithMessage(domain.ErrUnauthenticated, "wrong email")
		case err != nil:
			return errors.WithMessage(err, "get user by email")
		case user.SudirUserId != nil:
			return domain.ErrSudirAuthorization
		}

		if user.Blocked {
			a.auditService.SaveAuditAsync(ctx, user.Id, "Неуспешный вход. Пользователь заблокирован", entity.EventErrorLogin)
			return errors.WithMessagef(domain.ErrUnauthenticated, "user '%d' is blocked", user.Id)
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
		if err != nil {
			a.auditService.SaveAuditAsync(ctx, user.Id, "Неуспешный вход. Неверный пароль", entity.EventErrorLogin)
			return errors.WithMessage(domain.ErrUnauthenticated, "wrong password")
		}

		tokenString, expired, err = a.tokenService.GenerateToken(ctx, tx, user.Id)
		if err != nil {
			return errors.WithMessage(err, "generate token")
		}

		lastActiveAt := time.Now().UTC()
		err = tx.UpdateLastActiveAt(ctx, user.Id, lastActiveAt)
		if err != nil {
			return errors.WithMessage(err, "update user last_active_at")
		}

		a.auditService.SaveAuditAsync(ctx, user.Id, "Успешный вход через форму входа", entity.EventSuccessLogin)

		return nil
	})
	if err != nil {
		return nil, errors.WithMessage(err, "auth transaction")
	}

	return &domain.LoginResponse{
		Token:      tokenString,
		Expired:    expired,
		HeaderName: domain.AdminAuthHeaderName,
	}, nil
}

func (a Auth) LoginWithSudir(ctx context.Context, request domain.LoginSudirRequest) (*domain.LoginResponse, error) {
	var (
		user *entity.User

		tokenString string
		expired     string
	)

	err := a.txRunner.AuthTransaction(ctx, func(ctx context.Context, tx AuthTransaction) error {
		sudirUser, err := a.sudirService.Authenticate(ctx, request.AuthCode, tx)

		var authErr *entity.SudirAuthError
		switch {
		case errors.As(err, &authErr):
			a.logger.Error(ctx, "sudir authenticate: error occurred", log.String("error", err.Error()))
			return domain.ErrUnauthenticated
		case err != nil:
			return errors.WithMessage(err, "sudir authenticate")
		case sudirUser.SudirUserId == "" || sudirUser.Email == "":
			a.logger.Error(ctx, "sudir authenticate: missing sudirUserId or email",
				log.String("userId", sudirUser.SudirUserId),
				log.String("email", sudirUser.Email))
			return domain.ErrUnauthenticated
		}

		user, err = tx.UpsertBySudirUserId(ctx, entity.User{
			SudirUserId: &sudirUser.SudirUserId,
			FirstName:   sudirUser.FirstName,
			LastName:    sudirUser.LastName,
			FullName:    sudirUser.FullName,
			Email:       sudirUser.Email,
			Password:    "",
			Blocked:     false,
			UpdatedAt:   time.Now().UTC(),
			CreatedAt:   time.Now().UTC(),
		})
		if errors.Is(err, domain.ErrUserIsBlocked) {
			return errors.Errorf("user with sudir user id = %s is blocked", sudirUser.SudirUserId)
		}
		if err != nil {
			return errors.WithMessage(err, "upsert by sudir user id")
		}

		err = tx.UpsertUserRoleLinks(ctx, int(user.Id), sudirUser.RoleIds)
		if err != nil {
			return errors.WithMessage(err, "upsert user role links")
		}

		tokenString, expired, err = a.tokenService.GenerateToken(ctx, tx, user.Id)
		if err != nil {
			return errors.WithMessage(err, "generate token")
		}

		lastActiveAt := time.Now().UTC()
		err = tx.UpdateLastActiveAt(ctx, user.Id, lastActiveAt)
		if err != nil {
			return errors.WithMessage(err, "update user last_active_at")
		}

		return nil
	})
	if err != nil {
		return nil, errors.WithMessage(err, "auth transaction")
	}

	a.auditService.SaveAuditAsync(ctx, user.Id, "Успешный вход через СУДИР", entity.EventSuccessLogin)

	return &domain.LoginResponse{
		Token:      tokenString,
		Expired:    expired,
		HeaderName: domain.AdminAuthHeaderName,
	}, nil
}

func (a Auth) Logout(ctx context.Context, adminId int64, request *domain.LogoutRequest) error {
	err := a.txRunner.AuthTransaction(ctx, func(ctx context.Context, tx AuthTransaction) error {
		err := a.tokenService.RevokeAllByUserId(ctx, adminId)
		if err != nil {
			return errors.WithMessage(err, "revoke all tokens by user id")
		}

		lastActiveAt := time.Now().UTC()
		err = tx.UpdateLastActiveAt(ctx, adminId, lastActiveAt)
		if err != nil {
			return errors.WithMessage(err, "update user last_active_at")
		}

		if request != nil && request.Reason == "idle" {
			a.auditService.SaveAuditAsync(ctx, adminId, "Выход по бездействию", entity.EventSuccessLogout)
			return nil
		}

		a.auditService.SaveAuditAsync(ctx, adminId, "Выход", entity.EventSuccessLogout)
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "auth transaction")
	}

	return nil
}
