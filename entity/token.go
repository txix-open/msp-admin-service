package entity

import (
	"time"
)

const (
	TokenStatusAllowed = "ALLOWED"
	TokenStatusRevoked = "REVOKED"
	TokenStatusExpired = "EXPIRED"
)

type Token struct {
	Id        int
	Token     string
	UserId    int64
	Status    string
	ExpiredAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
