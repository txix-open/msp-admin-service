package entity

import "time"

type User struct {
	SudirUserId *string
	Id          int64
	RoleId      int
	FirstName   string
	LastName    string
	Email       string
	Password    string
	UpdatedAt   time.Time
	CreatedAt   time.Time
}

type SudirUser struct {
	RoleId      int
	SudirUserId string
	FirstName   string
	LastName    string
	Email       string
}

type CreateUser struct {
	RoleId    int
	FirstName string
	LastName  string
	Email     string
	Password  string
}

type UpdateUser struct {
	RoleId    int
	FirstName string
	LastName  string
	Email     string
	Password  string
}
