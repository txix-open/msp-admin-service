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

type SessionPageRequest struct {
	LimitOffestParams

	Order *OrderParams
	Query *SessionQuery
}

type SessionQuery struct {
	Id        *int
	UserId    *int
	Status    *string
	CreatedAt *DateFromToParams
	ExpiredAt *DateFromToParams
}

type SessionResponse struct {
	TotalCount int
	Items      []Session
}

type RevokeRequest struct {
	Id int `validate:"required"`
}
