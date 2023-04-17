package domain

import (
	"time"
)

type Audit struct {
	Id        int
	UserId    int
	Message   string
	CreatedAt time.Time
}

type AuditResponse struct {
	TotalCount int
	Items      []Audit
}

type PageRequest struct {
	Limit  int `valid:"required"`
	Offset int
}
