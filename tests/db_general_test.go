package tests_test

import (
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/db/query"
	"github.com/txix-open/isp-kit/test/dbt"
	"golang.org/x/crypto/bcrypt"
	"msp-admin-service/entity"
)

//nolint:gomnd
func InsertUser(db *dbt.TestDb, user entity.User) int64 {
	if user.Password != "" {
		passwordBytes, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
		user.Password = string(passwordBytes)
	}
	var id int64
	db.Must().SelectRow(&id, `insert into users (first_name, last_name, email, password, blocked, last_active_at)
	values($1,$2,$3,$4,$5,$6) returning id`,
		user.FirstName, user.LastName, user.Email, user.Password, user.Blocked, user.LastActiveAt)
	return id
}

func InsertSudirUser(db *dbt.TestDb, user entity.SudirUser) int64 {
	q, args, err := query.New().
		Insert("users").
		Columns("sudir_user_id", "first_name", "last_name", "email").
		Values(user.SudirUserId, user.FirstName, user.LastName, user.Email).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		panic(errors.WithMessagef(err, "insert sudir user"))
	}
	var id int64
	db.Must().SelectRow(&id, q, args...)
	return id
}

func InsertRole(db *dbt.TestDb, role entity.Role) int64 {
	if role.Permissions == nil {
		role.Permissions = []string{}
	}
	q, args, err := query.New().
		Insert("roles").
		Columns("name", "permissions").
		Values(role.Name, role.Permissions).
		Suffix("ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name, permissions = EXCLUDED.permissions RETURNING id").
		ToSql()
	if err != nil {
		panic(errors.WithMessagef(err, "insert role"))
	}
	var id int64
	db.Must().SelectRow(&id, q, args...)
	return id
}

func InsertUserRole(db *dbt.TestDb, role entity.UserRole) {
	q, args, err := query.New().
		Insert("user_roles").
		Columns("user_id", "role_id").
		Values(role.UserId, role.RoleId).
		ToSql()
	if err != nil {
		panic(errors.WithMessage(err, "insert user role"))
	}
	db.Must().Exec(q, args...)
}

func SelectTokenEntityByToken(db *dbt.TestDb, token string) entity.Token {
	tokenInfo := entity.Token{}
	db.Must().SelectRow(&tokenInfo,
		`SELECT token, user_id, status, expired_at, created_at, updated_at
					FROM tokens
					WHERE token = $1;`,
		token,
	)
	return tokenInfo
}

func InsertTokenEntity(db *dbt.TestDb, token entity.Token) {
	db.Must().ExecNamed(
		`
	INSERT INTO tokens
		(token, user_id, status, expired_at, created_at, updated_at)
		VALUES (:token, :user_id, :status, :expired_at, :created_at, :updated_at)
	`,
		token,
	)
}
