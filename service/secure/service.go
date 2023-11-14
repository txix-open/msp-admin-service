package secure

import (
	"context"
	"slices"
	"time"

	"github.com/pkg/errors"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
)

type TokenRep interface {
	Get(ctx context.Context, token string) (*entity.Token, error)
}

type UserRoleRepo interface {
	GetRoleEntitiesByUserId(ctx context.Context, userId int) ([]entity.Role, error)
}

type Service struct {
	tokenRep     TokenRep
	userRoleRepo UserRoleRepo
}

func NewService(tokenRep TokenRep, userRoleRepo UserRoleRepo) Service {
	return Service{
		tokenRep:     tokenRep,
		userRoleRepo: userRoleRepo,
	}
}

func (s Service) Authenticate(ctx context.Context, token string) (int64, error) {
	tokenInfo, err := s.tokenRep.Get(ctx, token)
	if err != nil {
		return 0, errors.WithMessage(err, "get token entity")
	}

	if time.Now().UTC().After(tokenInfo.ExpiredAt) ||
		tokenInfo.Status != entity.TokenStatusAllowed {
		return 0, domain.ErrTokenExpired
	}

	return tokenInfo.UserId, nil
}

func (s Service) Authorize(ctx context.Context, adminId int, permission string) (bool, error) {
	roles, err := s.userRoleRepo.GetRoleEntitiesByUserId(ctx, adminId)
	if err != nil {
		return false, errors.WithMessage(err, "get role entities by user id")
	}

	for _, role := range roles {
		exist := slices.Contains(role.Permissions, permission)
		if exist {
			return true, nil
		}
	}

	return false, nil
}
