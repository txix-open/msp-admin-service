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
	LastSessionCreatedAt time.Time
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
	FirstName   string
	LastName    string
	Email       string `valid:"required"`
	Role        string
	Roles       []int
	Permissions []string
}

type CreateUserRequest struct {
	Roles                []int
	FirstName            string
	LastName             string
	Email                string `valid:"required"`
	Password             string `valid:"required"`
	Description          string
	LastSessionCreatedAt time.Time
}

type UpdateUserRequest struct {
	Id                   int64 `valid:"required"`
	Roles                []int
	FirstName            string
	LastName             string
	Email                string
	Description          string
	Blocked              bool
	LastSessionCreatedAt time.Time
}

type DeleteResponse struct {
	Deleted int
}

type IdentitiesRequest struct {
	Ids []int64 `valid:"required"`
}

type IdRequest struct {
	UserId int `valid:"required"`
}
