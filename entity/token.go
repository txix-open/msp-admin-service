package entity

import (
	"time"
)

const (
	TokenStatusAllowed = "ALLOWED"
	TokenStatusRevoked = "REVOKED"
)

type Token struct {
	Token     string
	UserId    int64
	Status    string
	ExpiredAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
