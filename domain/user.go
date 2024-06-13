package domain

import (
	"time"
)

type User struct {
	Id                   int64
	Roles                []int
	FirstName            string
	LastName             string
	Email                string
	Description          string
	Blocked              bool
	LastSessionCreatedAt *time.Time
	UpdatedAt            time.Time
	CreatedAt            time.Time
}

type UsersResponse struct {
	Items []User
}

type UsersRequest struct {
	Ids    []int64
	Offset int
	Limit  int
	Email  string
}

type AdminUserShort struct {
	FirstName     string
	LastName      string
	Email         string `validate:"required"`
	Role          string
	Roles         []int
	Permissions   []string
	IdleTimeoutMs int
}

type CreateUserRequest struct {
	Roles       []int
	FirstName   string
	LastName    string
	Email       string `validate:"required"`
	Password    string `validate:"required"`
	Description string
}

type UpdateUserRequest struct {
	Id          int64 `validate:"required"`
	Roles       []int
	FirstName   string
	LastName    string
	Email       string
	Description string
	Blocked     bool
}

type DeleteResponse struct {
	Deleted int
}

type IdentitiesRequest struct {
	Ids []int64 `validate:"required"`
}

type IdRequest struct {
	UserId int `validate:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `validate:"required"`
	NewPassword string `validate:"required"`
}
