package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"msp-admin-service/domain"
	"msp-admin-service/entity"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type TokenRep interface {
	TokenSaver
	Get(ctx context.Context, token string) (*entity.Token, error)
	RevokeByUserId(ctx context.Context, userId int64, updatedAt time.Time) error
	AllByRequest(ctx context.Context, req domain.SessionPageRequest) ([]entity.Token, error)
	Count(ctx context.Context, reqQuery *domain.SessionQuery) (int64, error)
	UpdateStatus(ctx context.Context, id int, status string) error
}

type TokenSaver interface {
	Save(ctx context.Context, token entity.Token) error
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

func (s Token) GenerateToken(ctx context.Context, repo TokenSaver, id int64) (string, string, error) {
	cryptoRand := make([]byte, 128) //nolint:mnd
	_, err := rand.Read(cryptoRand)
	if err != nil {
		return "", "", errors.WithMessage(err, "crypto/rand read")
	}
	random := hex.EncodeToString(cryptoRand)

	createdAt := time.Now().UTC()
	expiredAt := createdAt.Add(s.lifeTime)

	err = repo.Save(ctx, entity.Token{
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

func (s Token) RevokeAllByUserId(ctx context.Context, userId int64) error {
	updatedAt := time.Now().UTC()
	err := s.tokenRep.RevokeByUserId(ctx, userId, updatedAt)
	if err != nil {
		return errors.WithMessage(err, "set revoked status")
	}

	return nil
}

func (s Token) All(ctx context.Context, req domain.SessionPageRequest) (*domain.SessionResponse, error) {
	var tokens []entity.Token
	var total int64

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		tokens, err = s.tokenRep.AllByRequest(ctx, req)
		if err != nil {
			return errors.WithMessage(err, "get all tokens")
		}
		return nil
	})
	group.Go(func() error {
		var err error
		total, err = s.tokenRep.Count(ctx, req.Query)
		if err != nil {
			return errors.WithMessage(err, "count all tokens")
		}
		return nil
	})
	err := group.Wait()
	if err != nil {
		return nil, errors.WithMessage(err, "wait workers")
	}

	items := make([]domain.Session, 0)
	for _, token := range tokens {
		items = append(items, domain.Session{
			Id:        token.Id,
			UserId:    int(token.UserId),
			Status:    token.Status,
			ExpiredAt: token.ExpiredAt,
			CreatedAt: token.CreatedAt,
		})
	}

	result := domain.SessionResponse{
		TotalCount: int(total),
		Items:      items,
	}

	return &result, nil
}

func (s Token) Revoke(ctx context.Context, id int) error {
	err := s.tokenRep.UpdateStatus(ctx, id, entity.TokenStatusRevoked)
	if err != nil {
		return errors.WithMessage(err, "token update status")
	}
	return nil
}
