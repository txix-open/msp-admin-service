package entity

import "time"

type User struct {
	SudirUserId *string
	Id          int64
	FirstName   string
	LastName    string
	Description string
	Email       string
	Password    string
	Blocked     bool
	UpdatedAt   time.Time
	CreatedAt   time.Time
}

type SudirUser struct {
	RoleIds     []int
	SudirUserId string
	FirstName   string
	LastName    string
	Email       string
	Description string
}

type UpdateUser struct {
	FirstName   string
	LastName    string
	Email       string
	Description string
}
