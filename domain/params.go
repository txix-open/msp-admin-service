package domain

import "time"

type LimitOffestParams struct {
	Limit  int `validate:"required"`
	Offset int
}

type OrderParams struct {
	Field string
	Type  string `validate:"oneof=asc desc ASC DESC"`
}

type DateFromToParams struct {
	From time.Time `validate:"required"`
	To   time.Time `validate:"required"`
}
