package repository

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"msp-admin-service/domain"
	"msp-admin-service/entity"

	"github.com/Masterminds/squirrel"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/db/query"
	"github.com/txix-open/isp-kit/metrics/sql_metrics"
)

type User struct {
	db db.DB
}

func NewUser(db db.DB) User {
	return User{db: db}
}

func (u User) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "User.GetUserByEmail")

	q, args, err := query.New().
		Select("*").
		From("users").
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	user := entity.User{}
	err = u.db.SelectRow(ctx, &user, q, args...)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, domain.ErrNotFound
	case err != nil:
		return nil, errors.WithMessage(err, "db select")
	default:
		return &user, nil
	}
}

func (u User) GetUserById(ctx context.Context, identity int64) (*entity.User, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "User.GetUserById")

	q, args, err := query.New().
		Select("*").
		From("users").
		Where(squirrel.Eq{"id": identity}).
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	user := entity.User{}
	err = u.db.SelectRow(ctx, &user, q, args...)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, domain.ErrNotFound
	case err != nil:
		return nil, errors.WithMessage(err, "db select")
	default:
		return &user, nil
	}
}

func (u User) GetUserByEmailAndSudirId(ctx context.Context, email string, sudirUserId string) (*entity.User, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "User.GetUserByEmail")

	equalClause := squirrel.Eq{"email": email, "sudir_user_id": nil}
	if sudirUserId != "" {
		equalClause["sudir_user_id"] = sudirUserId
	}

	q, args, err := query.New().
		Select("*").
		From("users").
		Where(equalClause).
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	user := entity.User{}
	err = u.db.SelectRow(ctx, &user, q, args...)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, domain.ErrNotFound
	case err != nil:
		return nil, errors.WithMessage(err, "db select")
	default:
		return &user, nil
	}
}

func (u User) UpsertBySudirUserId(ctx context.Context, user entity.User) (*entity.User, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "User.UpsertBySudirUserId")

	selectQ := `
	insert into users as u (first_name, last_name, email, created_at, updated_at, sudir_user_id) 
	values ($1, $2, $3, $4, $5, $6)
    on conflict (sudir_user_id) do update 
    set first_name = excluded.first_name,
        last_name = excluded.last_name,
        email = excluded.email,
        updated_at = excluded.updated_at
    where u.blocked = false
    returning *
`
	result := entity.User{}
	err := u.db.SelectRow(ctx,
		&result,
		selectQ,
		user.FirstName, user.LastName, user.Email, user.CreatedAt, user.UpdatedAt, user.SudirUserId,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserIsBlocked
	}
	if err != nil {
		return nil, errors.WithMessagef(err, "select row: %s", selectQ)
	}

	return &result, nil
}

func (u User) GetUsers(ctx context.Context, req domain.UsersPageRequest) ([]entity.User, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "User.GetUsers")

	q := query.New().
		Select("*", "(SELECT max(created_at) FROM tokens WHERE tokens.user_id = users.id) as last_session_created_at").
		From("users").
		Offset(req.Offset).
		Limit(req.Limit)

	if req.Order.Field == "userId" {
		q = q.OrderBy("last_name "+req.Order.Type, "first_name "+req.Order.Type)
	} else {
		q = q.OrderBy(strcase.ToSnake(req.Order.Field) + " " + req.Order.Type)
	}

	query, args, err := reqUsersQuery(q, req.Query).ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	users := make([]entity.User, 0)
	err = u.db.Select(ctx, &users, query, args...)
	if err != nil {
		return nil, errors.WithMessage(err, "db select")
	}

	return users, nil
}

func (u User) GetUsersByEmail(ctx context.Context, email string) ([]entity.User, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "User.GetUserByEmail")

	q, args, err := query.New().
		Select("*").
		From("users").
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	users := make([]entity.User, 0)
	err = u.db.Select(ctx, &users, q, args...)

	switch {
	case errors.Is(err, sql.ErrNoRows) || len(users) == 0:
		return nil, domain.ErrNotFound
	case err != nil:
		return nil, errors.WithMessage(err, "db select")
	default:
		return users, nil
	}
}

func (u User) Insert(ctx context.Context, user entity.User) (int, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "User.Insert")

	insertQ, args, err := query.New().
		Insert("users").
		Columns("first_name", "last_name", "description",
			"email", "password", "created_at", "updated_at").
		Values(user.FirstName, user.LastName, user.Description,
			user.Email, user.Password, user.CreatedAt, user.UpdatedAt).
		Suffix("returning id").
		ToSql()
	if err != nil {
		return 0, errors.WithMessage(err, "build query")
	}

	id := 0
	err = u.db.SelectRow(ctx, &id, insertQ, args...)
	if err != nil {
		return 0, errors.WithMessagef(err, "select row: %s", insertQ)
	}

	return id, nil
}

func (u User) UpdateUser(ctx context.Context, id int64, user entity.UpdateUser) (*entity.User, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "User.UpdateUser")

	// return every except password
	q, args, err := query.New().
		Update("users").
		SetMap(map[string]interface{}{
			"first_name":  user.FirstName,
			"last_name":   user.LastName,
			"email":       user.Email,
			"description": user.Description,
		}).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING id, first_name, last_name, email, sudir_user_id, description, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	returning := entity.User{}
	err = u.db.SelectRow(ctx, &returning, q, args...)
	if err != nil {
		return nil, errors.WithMessage(err, "db select")
	}

	return &returning, nil
}

func (u User) DeleteUser(ctx context.Context, ids []int64) (int, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "User.DeleteUser")

	q, args, err := query.New().
		Delete("users").
		Where(squirrel.Eq{"id": ids}).
		ToSql()
	if err != nil {
		return 0, errors.WithMessage(err, "build query")
	}

	result, err := u.db.Exec(ctx, q, args...)
	if err != nil {
		return 0, errors.WithMessage(err, "db exec")
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, errors.WithMessage(err, "get row affected")
	}

	return int(affected), err
}

func (u User) ChangeBlockStatus(ctx context.Context, userId int) (bool, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "User.ChangeBlockStatus")

	query := "update users set blocked = not blocked where id = $1 returning blocked"
	blocked := false
	err := u.db.SelectRow(ctx, &blocked, query, userId)
	if err != nil {
		return false, errors.WithMessagef(err, "select row: %s", query)
	}
	return blocked, nil
}

func (u User) Block(ctx context.Context, userId int) (*entity.User, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "User.Block")

	var user entity.User
	query := "update users set blocked = true where id = $1 returning *"
	err := u.db.SelectRow(ctx, &user, query, userId)
	if err != nil {
		return nil, errors.WithMessagef(err, "select row: %s", query)
	}
	return &user, nil
}

func (u User) LastAccessNotBlockedUsers(ctx context.Context) (map[int64]time.Time, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "User.LastAccessNotBlockedUsers")

	q, args, err := query.New().
		Select("id, last_active_at").
		From("users").
		Where(squirrel.Eq{"blocked": false}).
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	users := make([]entity.User, 0)
	err = u.db.Select(ctx, &users, q, args...)
	if err != nil {
		return nil, errors.WithMessagef(err, "select query: %s", q)
	}

	result := make(map[int64]time.Time, 0)
	for _, user := range users {
		result[user.Id] = user.LastActiveAt
	}

	return result, nil
}

func (u User) ChangePassword(ctx context.Context, userId int64, newPassword string) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "User.ChangePassword")

	q, args, err := query.New().
		Update("users").
		Where(squirrel.Eq{"id": userId}).
		Set("password", newPassword).
		ToSql()
	if err != nil {
		return errors.WithMessage(err, "user.repo.ChangePassword: build query")
	}

	_, err = u.db.Exec(ctx, q, args...)
	if err != nil {
		return errors.WithMessagef(err, "user.repo.ChangePassword: exec query: %s", q)
	}

	return nil
}

func (u User) UpdateLastActiveAt(ctx context.Context, userId int64, lastActiveAt time.Time) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "User.UpdateLastActiveAt")

	q := "update users set last_active_at = $1 where id = $2"
	_, err := u.db.Exec(ctx, q, lastActiveAt, userId)
	if err != nil {
		return errors.WithMessagef(err, "user.repo.UpdateLastActiveAt: exec query: %s", q)
	}
	return nil
}

func reqUsersQuery(q squirrel.SelectBuilder, reqQuery *domain.UserQuery) squirrel.SelectBuilder {
	if reqQuery == nil {
		return q
	}

	if reqQuery.Id != nil {
		q = q.Where(squirrel.ILike{"id::text": strconv.Itoa(*reqQuery.Id) + "%"})
	}

	if reqQuery.UserId != nil { // поиск в ui по имени, но в бд - по id юзера
		q = q.Where("id = ?", *reqQuery.UserId)
	}

	if reqQuery.Description != nil {
		q = q.Where(squirrel.ILike{"description": "%" + *reqQuery.Description + "%"})
	}

	if reqQuery.Email != nil {
		q = q.Where(squirrel.ILike{"email": "%" + *reqQuery.Email + "%"})
	}

	return q
}
