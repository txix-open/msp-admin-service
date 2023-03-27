package repository

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/db/query"
	"github.com/pkg/errors"
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

func (u User) GetUserBySudirUserId(ctx context.Context, id string) (*entity.User, error) {
	q, args, err := query.New().
		Select("id", "role_id", "first_name", "last_name", "password",
			"email", "sudir_user_id", "created_at", "updated_at").
		From("users").
		Where(squirrel.Eq{"sudir_user_id": id}).
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

func (u User) GetUsers(ctx context.Context, ids []int64, offset, limit int, email string) ([]entity.User, error) {
	// take every except password
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

func (u User) CreateUser(ctx context.Context, user entity.CreateUser) (*entity.User, error) {
	// return every except password
	q, args, err := query.New().
		Insert("users").
		Columns("role_id", "first_name", "last_name", "email", "password").
		Values(user.RoleId, user.FirstName, user.LastName, user.Email, user.Password).
		Suffix("RETURNING id, role_id, first_name, last_name, email, sudir_user_id, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	returning := entity.User{}
	err = u.db.SelectRow(ctx, &returning, q, args...)
	if err != nil {
		return nil, errors.WithMessage(err, "db select")
	}

	return &returning, err
}

func (u User) CreateSudirUser(ctx context.Context, user entity.SudirUser) (*entity.User, error) {
	// return every except password
	q, args, err := query.New().
		Insert("users").
		Columns("role_id", "sudir_user_id", "first_name", "last_name", "email").
		Values(user.RoleId, user.SudirUserId, user.FirstName, user.LastName, user.Email).
		Suffix("RETURNING id, role_id, first_name, last_name, email, sudir_user_id, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	returning := entity.User{}
	err = u.db.SelectRow(ctx, &returning, q, args...)
	if err != nil {
		return nil, errors.WithMessage(err, "db select")
	}

	return &returning, err
}

func (u User) UpdateUser(ctx context.Context, id int64, user entity.UpdateUser) (*entity.User, error) {
	// return every except password
	q, args, err := query.New().
		Update("users").
		SetMap(map[string]interface{}{
			"role_id":    user.RoleId,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
			"password":   user.Password,
		}).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING id, role_id, first_name, last_name, email, sudir_user_id, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	returning := entity.User{}
	err = u.db.SelectRow(ctx, &returning, q, args...)
	if err != nil {
		return nil, errors.WithMessage(err, "db select")
	}

	return &returning, err
}

func (u User) DeleteUser(ctx context.Context, ids []int64) (int, error) {
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
