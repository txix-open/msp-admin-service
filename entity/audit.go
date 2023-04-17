package entity

import (
	"time"
)

type Audit struct {
	Id        int
	UserId    int
	Message   string
	CreatedAt time.Time
}
