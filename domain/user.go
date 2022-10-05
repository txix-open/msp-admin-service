package domain

import (
	"time"
)

type User struct {
	SudirUserId *string `json:",omitempty"`
	Id          int64
	RoleId      int
	FirstName   string
	LastName    string
	Email       string
	Password    string `json:",omitempty"`
	UpdatedAt   time.Time
	CreatedAt   time.Time
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
	FirstName string
	LastName  string
	Email     string `valid:"required"`
	Role      string
}

type CreateUserRequest struct {
	RoleId    int
	FirstName string
	LastName  string
	Email     string `valid:"required"`
	Password  string `valid:"required"`
}

type UpdateUserRequest struct {
	Id        int64 `valid:"required"`
	RoleId    int
	FirstName string
	LastName  string
	Email     string
	Password  string
}

type DeleteResponse struct {
	Deleted int
}

type IdentitiesRequest struct {
	Ids []int64 `valid:"required"`
}
