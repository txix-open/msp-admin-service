package entity

import (
	"time"

	"github.com/integration-system/isp-kit/json"
)

type Role struct {
	Id            int
	Name          string
	ExternalGroup string
	ChangeMessage string
	Permissions   PermList
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PermList []string

func (p *PermList) Scan(src any) error {
	var data []byte
	switch v := src.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	}

	return json.Unmarshal(data, p)
}
