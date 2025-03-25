package entity

import (
	"database/sql/driver"
	"time"

	"github.com/txix-open/isp-kit/json"
)

const (
	GroupOperationAdd    = "ADD"
	GroupOperationDelete = "DELETE"
)

type Role struct {
	Id            int
	Name          string
	ExternalGroup string
	Permissions   PermList
	Immutable     bool
	Exclusive     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PermList []string

// nolint
func (p *PermList) Scan(src any) error {
	return json.Unmarshal(src.([]byte), p)
}

func (p *PermList) Value() (driver.Value, error) {
	bytes, err := json.Marshal(p)
	return driver.Value(bytes), err
}
