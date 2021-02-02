package entity

import "time"

type AdminToken struct {
	tableName string     `pg:"?db_schema.tokens" json:"-"`
	Id        int64      `json:"id"`
	UserId    int64      `json:"userId"`
	Token     string     `json:"token"`
	ExpiredAt *time.Time `json:"expiredAt"`
	CreatedAt time.Time  `json:"createdAt" sql:",null"`
}

type AdminUser struct {
	tableName string    `pg:"?db_schema.users" json:"-"`
	Id        int64     `json:"id"`
	Image     string    `json:"image"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email" valid:"required~Required"`
	Password  string    `json:"password,omitempty"`
	Phone     string    `json:"phone"`
	UpdatedAt time.Time `json:"updatedAt" sql:",null"`
	CreatedAt time.Time `json:"createdAt" sql:",null"`
}
