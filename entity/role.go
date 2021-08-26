package entity

import (
	"time"
)

type Role struct {
	tableName   string `pg:"?db_schema.roles" json:"-"`
	Id          int
	Name        string
	Rights      interface{}
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
