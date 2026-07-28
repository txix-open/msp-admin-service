package domain

import "time"

const (
	DefaultOrderField = "created_at"
	DefaultOrderType  = "desc"
)

type LimitOffestParams struct {
	Limit  uint64 `validate:"required"` // Максимальное количество записей
	Offset uint64 // Отступ для выборки
}

type OrderParams struct {
	Field string // Поле сортировки
	Type  string `validate:"oneof=asc desc ASC DESC"` // Порядок сортировки
}

type DateFromToParams struct {
	From time.Time `validate:"required"` // Дата с
	To   time.Time `validate:"required"` // Дата до
}
