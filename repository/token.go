package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/integration-system/isp-kit/db"
	"github.com/pkg/errors"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
)

type Token struct {
	db db.DB
}

func NewToken(db db.DB) Token {
	return Token{
		db: db,
	}
}

func (r Token) Save(ctx context.Context, token entity.Token) error {
	q := `
	INSERT INTO tokens
		(token, user_id, status, expired_at, created_at, updated_at)
		VALUES (:token, :user_id, :status, :expired_at, :created_at, :updated_at)
	`
	_, err := r.db.ExecNamed(ctx, q, token)
	if err != nil {
		return errors.WithMessage(err, "save token row")
	}

	return nil
}

func (r Token) GetEntity(ctx context.Context, token string) (*entity.Token, error) {
	result := entity.Token{}
	q := `
	SELECT token, user_id, status, expired_at, created_at, updated_at 
		FROM tokens
		WHERE token = $1;
	`
	err := r.db.SelectRow(ctx, &result, q, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, errors.WithMessage(err, "select token row by token")
	}

	return &result, nil
}

func (r Token) RevokeByUserId(ctx context.Context, userId int64, updatedAt time.Time) error {
	q := `
	UPDATE tokens
		SET status = $1, updated_at = $2
		WHERE user_id = $3 AND status = $4;
	`
	_, err := r.db.Exec(ctx, q, entity.TokenStatusRevoked, updatedAt, userId, entity.TokenStatusAllowed)
	if err != nil {
		return errors.WithMessage(err, "update token status")
	}

	return nil
}
