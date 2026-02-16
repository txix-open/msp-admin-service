package domain

import "time"

const (
	DefaultOrderField = "created_at"
	DefaultOrderType  = "desc"
)

type LimitOffestParams struct {
	Limit  uint64 `validate:"required"`
	Offset uint64
}

type OrderParams struct {
	Field string
	Type  string `validate:"oneof=asc desc ASC DESC"`
}

type DateFromToParams struct {
	From time.Time `validate:"required"`
	To   time.Time `validate:"required"`
}
