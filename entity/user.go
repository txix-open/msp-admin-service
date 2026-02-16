package entity

import "time"

type User struct {
	SudirUserId          *string
	Id                   int64
	FirstName            string
	LastName             string
	FullName             string
	Description          string
	Email                string
	Password             string
	Blocked              bool
	LastActiveAt         time.Time
	UpdatedAt            time.Time
	CreatedAt            time.Time
	LastSessionCreatedAt *time.Time
}

type SudirUser struct {
	RoleIds     []int
	SudirUserId string
	FirstName   string
	LastName    string
	Email       string
	Description string
	FullName    string
}

type UpdateUser struct {
	FirstName   string
	LastName    string
	FullName    string
	Email       string
	Description string
}
