package domain

import (
	"time"
)

type Role struct {
	Id            int      // Идентификатор роли
	Name          string   // Название роли
	ExternalGroup string   // Внешнее название группы
	ChangeMessage string   // Причина изменений
	Permissions   []string // Набор прав
	Immutable     bool
	Exclusive     bool
	CreatedAt     time.Time // Дата создания
	UpdatedAt     time.Time // Дата обновления
}

type CreateRoleRequest struct {
	Name          string   // Название роли
	ExternalGroup string   // Внешнее название группы
	ChangeMessage string   // Причина изменений
	Permissions   []string // Набор прав
}

type UpdateRoleRequest struct {
	Id            int      // Идентификатор роли
	Name          string   // Название роли
	ExternalGroup string   // Внешнее название группы
	ChangeMessage string   // Причина изменений
	Permissions   []string // Набор прав
}

type DeleteRoleRequest struct {
	Id int // Идентификатор роли
}
