package domain

import (
	"time"
)

type Session struct {
	Id        int
	UserId    int
	Status    string
	ExpiredAt time.Time
	CreatedAt time.Time
}

type SessionResponse struct {
	TotalCount int
	Items      []Session
}

type SessionRequest struct {
	Limit  int
	Offset int
}

type RevokeRequest struct {
	Id int `validate:"required"`
}
