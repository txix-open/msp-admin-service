package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"

	"msp-admin-service/domain"
	"msp-admin-service/entity"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/db/query"
	"github.com/txix-open/isp-kit/metrics/sql_metrics"
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
	ctx = sql_metrics.OperationLabelToContext(ctx, "Token.Save")

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

func (r Token) Get(ctx context.Context, token string) (*entity.Token, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Token.Get")

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
	ctx = sql_metrics.OperationLabelToContext(ctx, "Token.RevokeByUserId")

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

//nolint:dupl,gosec
func (r Token) All(ctx context.Context, req domain.SessionPageRequest) ([]entity.Token, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Token.All")

	if req.Order == nil {
		req.Order = &domain.OrderParams{
			Field: domain.DefaultOrderField,
			Type:  domain.DefaultOrderType,
		}
	}

	q := query.New().
		Select("*").
		From("tokens").
		OrderBy(req.Order.Field + " " + req.Order.Type).
		Offset(uint64(req.Offset)).
		Limit(uint64(req.Limit))

	query, args, err := reqTokenQuery(q, req.Query).ToSql()
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
	ctx = sql_metrics.OperationLabelToContext(ctx, "Token.UpdateStatus")

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
	ctx = sql_metrics.OperationLabelToContext(ctx, "Token.Count")

	count := int64(0)
	err := r.db.SelectRow(ctx, &count, "select count(*) from tokens")
	if err != nil {
		return 0, errors.WithMessage(err, "select count")
	}
	return count, nil
}

func (r Token) UpdateStatusByUserId(ctx context.Context, userId int, status string) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Token.UpdateStatusByUserId")

	q := `
	UPDATE tokens SET status = $1, updated_at = $2 where user_id = $3;
`
	_, err := r.db.Exec(ctx, q, status, time.Now().UTC(), userId)
	if err != nil {
		return errors.WithMessage(err, "update token status")
	}

	return nil
}

func (r Token) LastAccessByUserIds(ctx context.Context, userIds []int, reqQuery *domain.UserQuery) (map[int64]*time.Time, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Token.LastAccessNotBlockedUsers")

	q := query.New().
		Select("user_id", "max(created_at) as created_at").
		From("tokens").
		Where(squirrel.Eq{"user_id": userIds}).
		GroupBy("user_id")

	if reqQuery != nil {
		if reqQuery.LastSessionCreatedAt != nil {
			q = q.Where(squirrel.GtOrEq{"created_at": reqQuery.LastSessionCreatedAt.From}).
				Where(squirrel.LtOrEq{"created_at": reqQuery.LastSessionCreatedAt.To})
		}
	}

	query, args, err := q.ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	tokens := make([]entity.Token, 0)
	err = r.db.Select(ctx, &tokens, query, args...)
	if err != nil {
		return nil, errors.WithMessagef(err, "select query: %s", query)
	}

	result := make(map[int64]*time.Time)
	for _, token := range tokens {
		createdAt := token.CreatedAt
		result[token.UserId] = &createdAt
	}

	return result, nil
}

func reqTokenQuery(q squirrel.SelectBuilder, reqQuery *domain.SessionQuery) squirrel.SelectBuilder {
	if reqQuery == nil {
		return q
	}

	switch {
	case reqQuery.Id != nil:
		{
			q = q.Where("id = ?", *reqQuery.Id)
		}
	case reqQuery.UserId != nil:
		{
			q = q.Where("user_id = ?", *reqQuery.UserId)
		}
	case reqQuery.Status != nil:
		{
			q = q.Where("status = ?", *reqQuery.Status)
		}
	case reqQuery.CreatedAt != nil:
		{
			q = q.Where(squirrel.GtOrEq{"created_at": reqQuery.CreatedAt.From}).
				Where(squirrel.LtOrEq{"created_at": reqQuery.CreatedAt.To})
		}
	case reqQuery.ExpiredAt != nil:
		{
			q = q.Where(squirrel.GtOrEq{"expired_at": reqQuery.ExpiredAt.From}).
				Where(squirrel.LtOrEq{"expired_at": reqQuery.ExpiredAt.To})
		}
	}

	return q
}
