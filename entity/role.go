package entity

import (
	"time"
)

type Role struct {
	Id          int
	Name        string
	Rights      interface{}
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
