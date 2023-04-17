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
	Blocked     bool
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

type UpdateUser struct {
	RoleId    int
	FirstName string
	LastName  string
	Email     string
	Password  string
}
