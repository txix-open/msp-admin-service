package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/db/query"
	"github.com/integration-system/isp-kit/metrics/sql_metrics"
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
	sql_metrics.OperationLabelToContext(ctx, "Token.Save")

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
	sql_metrics.OperationLabelToContext(ctx, "Token.GetEntity")

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
	sql_metrics.OperationLabelToContext(ctx, "Token.RevokeByUserId")

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

func (r Token) All(ctx context.Context, limit int, offset int) ([]entity.Token, error) {
	sql_metrics.OperationLabelToContext(ctx, "Token.All")

	query, args, err := query.New().
		Select("*").
		From("tokens").
		OrderBy("created_at DESC").
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	tokens := make([]entity.Token, 0)
	err = r.db.Select(ctx, &tokens, query, args...)
	if err != nil {
		return nil, errors.WithMessagef(err, "select query %s", query)
	}

	return tokens, nil
}

func (r Token) UpdateStatus(ctx context.Context, id int, status string) error {
	sql_metrics.OperationLabelToContext(ctx, "Token.UpdateStatus")

	q := `
	UPDATE tokens SET status = $1, updated_at = $2 where id = $3;
`
	_, err := r.db.Exec(ctx, q, status, time.Now().UTC(), id)
	if err != nil {
		return errors.WithMessage(err, "update token status")
	}

	return nil
}

func (r Token) Count(ctx context.Context) (int64, error) {
	sql_metrics.OperationLabelToContext(ctx, "Token.Count")

	count := int64(0)
	err := r.db.SelectRow(ctx, &count, "select count(*) from tokens")
	if err != nil {
		return 0, errors.WithMessage(err, "select count")
	}
	return count, nil
}

func (r Token) UpdateStatusByUserId(ctx context.Context, userId int, status string) error {
	sql_metrics.OperationLabelToContext(ctx, "Token.UpdateStatusByUserId")

	q := `
	UPDATE tokens SET status = $1, updated_at = $2 where user_id = $3;
`
	_, err := r.db.Exec(ctx, q, status, time.Now().UTC(), userId)
	if err != nil {
		return errors.WithMessage(err, "update token status")
	}

	return nil
}

func (r Token) LastAccessNotBlockedUsers(ctx context.Context) (map[int64]time.Time, error) {
	sql_metrics.OperationLabelToContext(ctx, "Token.LastAccessNotBlockedUsers")

	query := `
	select t.user_id, max(t.created_at) as created_at from tokens t 
	    join users u on t.user_id = u.id
	    where u.blocked = false
	    group by t.user_id
`
	tokens := make([]entity.Token, 0)
	err := r.db.Select(ctx, &tokens, query)
	if err != nil {
		return nil, errors.WithMessagef(err, "select query: %s", query)
	}

	result := make(map[int64]time.Time, 0)
	for _, token := range tokens {
		result[token.UserId] = token.CreatedAt
	}

	return result, nil
}
