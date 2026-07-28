package domain

import (
	"time"
)

type Session struct {
	Id        int       // Идентификатор сессии
	UserId    int       // Идентификатор пользоватеккля
	Status    string    // Статус сессии
	ExpiredAt time.Time // Дата протухания сессии
	CreatedAt time.Time // Дата создания сессии
}

type SessionPageRequest struct {
	LimitOffestParams

	Order *OrderParams  // Параметры сортировки
	Query *SessionQuery // Параметры выборки
}

type SessionQuery struct {
	Id        *int              // Фильтр по идентификатору сессии
	UserId    []int             // Фильтр по идентификаторам пользователей
	Status    []string          // Фильтр по статусам
	CreatedAt *DateFromToParams // Фильтр по дате создания
	ExpiredAt *DateFromToParams // Фильтр по дате протухания
}

type SessionResponse struct {
	TotalCount int       // Количество сессий
	Items      []Session // Список сессий
}

type RevokeRequest struct {
	Id int `validate:"required"` // Идентификатор сессии
}
