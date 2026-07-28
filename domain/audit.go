package domain

import (
	"time"
)

type Audit struct {
	Id        int       // Идентификатор audit'а
	UserId    int       // Идентификатор пользователя
	Message   string    // Сообщение
	CreatedAt time.Time // Дата события
}

type AuditResponse struct {
	TotalCount int     // Количество audit'ов
	Items      []Audit // Список audit'ов
}

type AuditPageRequest struct {
	LimitOffestParams

	Order *OrderParams // Параметры сортировки
	Query *AuditQuery  // Параметры выборки
}

type AuditQuery struct {
	Id        *int              // Фильтр по идентификатору audit
	UserId    []int             // Фильтр по идентификатору пользователей
	Message   *string           // Фильтр по сообщению
	CreatedAt *DateFromToParams // Фильтр по дате
}

type SetAuditEvent struct {
	Event   string // Название события
	Enabled bool   // Включено/выключено
}

type AuditEvent struct {
	Event   string // Название события
	Name    string //
	Enabled bool
}
