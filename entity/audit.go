package entity

import (
	"time"
)

type Audit struct {
	Id        int
	UserId    int
	Message   string
	Event     string
	CreatedAt time.Time
}
