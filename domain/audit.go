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
	Limit  int `validate:"required"`
	Offset int
}

type SetAuditEvent struct {
	Event   string
	Enabled bool
}

type AuditEvent struct {
	Event   string
	Name    string
	Enabled bool
}
