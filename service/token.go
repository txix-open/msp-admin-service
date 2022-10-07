package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/pkg/errors"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
)

type TokenRep interface {
	Save(ctx context.Context, token entity.Token) error
	GetEntity(ctx context.Context, token string) (*entity.Token, error)
}

type Token struct {
	tokenRep TokenRep
	lifeTime time.Duration
}

func NewToken(tokenRep TokenRep, lifeTimeInSec int) Token {
	return Token{
		lifeTime: time.Second * time.Duration(lifeTimeInSec),
		tokenRep: tokenRep,
	}
}

func (s Token) GenerateToken(ctx context.Context, id int64) (string, string, error) {
	cryptoRand := make([]byte, 128) //nolint:gomnd
	_, err := rand.Read(cryptoRand)
	if err != nil {
		return "", "", errors.WithMessage(err, "crypto/rand read")
	}
	random := hex.EncodeToString(cryptoRand)

	createdAt := time.Now().UTC()
	expiredAt := createdAt.Add(s.lifeTime)

	err = s.tokenRep.Save(ctx, entity.Token{
		Token:     random,
		UserId:    id,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: expiredAt,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})
	if err != nil {
		return "", "", errors.WithMessage(err, "save token")
	}

	return random, expiredAt.String(), nil
}

func (s Token) GetUserId(ctx context.Context, token string) (int64, error) {
	tokenInfo, err := s.tokenRep.GetEntity(ctx, token)
	if err != nil {
		return 0, errors.WithMessage(err, "get token entity")
	}

	if time.Now().UTC().After(tokenInfo.ExpiredAt) ||
		tokenInfo.Status != entity.TokenStatusAllowed {
		return 0, domain.ErrTokenExpired
	}

	return tokenInfo.UserId, nil
}
