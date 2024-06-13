package repository

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/db/query"
	"github.com/txix-open/isp-kit/metrics/sql_metrics"
	"msp-admin-service/domain"
	"msp-admin-service/entity"
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

func (u User) GetUsers(ctx context.Context, ids []int64, offset, limit int, email string) ([]entity.User, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "User.GetUsers")

	q := query.New().
		Select("*").
		From("users")

	if len(ids) > 0 {
		q = q.Where(squirrel.Eq{"id": ids})
	}
	if email != "" {
		q = q.Where("email LIKE ?", "%"+email+"%")
	}
	qstring, args, err := q.
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	users := make([]entity.User, 0)
	err = u.db.Select(ctx, &users, qstring, args...)
	if err != nil {
		return nil, errors.WithMessage(err, "db select")
	}

	return users, nil
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
		return errors.WithMessage(err, "user.repo.ChangePassword: exec query")
	}

	return nil
}
