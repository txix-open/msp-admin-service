package domain

import (
	"time"
)

type Role struct {
	Id            int
	Name          string
	ExternalGroup string
	ChangeMessage string
	Permissions   []string
	Immutable     bool
	Exclusive     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreateRoleRequest struct {
	Name          string
	ExternalGroup string
	ChangeMessage string
	Permissions   []string
}

type UpdateRoleRequest struct {
	Id            int
	Name          string
	ExternalGroup string
	ChangeMessage string
	Permissions   []string
}

type DeleteRoleRequest struct {
	Id int
}
