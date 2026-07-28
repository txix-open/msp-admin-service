package domain

import (
	"time"
)

type User struct {
	Id                   int64      // Идентификатор пользователя
	Roles                []int      // Список идентификаторов ролей
	FirstName            string     // Имя
	LastName             string     // Фамилия
	FullName             string     // ФИО
	Email                string     // Электронная почта
	Description          string     // Описание
	Blocked              bool       // Флаг блокировки
	LastSessionCreatedAt *time.Time // Дата создания последней сессии
	UpdatedAt            time.Time  // Дата обновлёния
	CreatedAt            time.Time  // Дата создания
}

type UsersResponse struct {
	Items []User // Список пользователей
}

type UserQuery struct {
	Id                   *int              // Фильтр по идентификатору пользователя
	UserId               []int             // Фильтр по идентификаторам пользователя
	Description          *string           // Фильтр по описанию
	Email                *string           // Фильтр по электроннйо почты
	Roles                []int             // Фильтр по ролям
	LastSessionCreatedAt *DateFromToParams // Фильтр по дате создания последней сессии
}

type UsersPageRequest struct {
	LimitOffestParams

	Order *OrderParams // Параметры сортировки
	Query *UserQuery   // Параметры выборки
}

type AdminUserShort struct {
	FirstName     string   // Имя
	LastName      string   // Фамилия
	FullName      string   // ФИО
	Email         string   `validate:"required"` // Электронный адрес
	Role          string   // Название роли
	Roles         []int    // Список идентификаторов ролей
	RoleNames     []string // Список названия ролей
	Permissions   []string // Список прав
	IdleTimeoutMs int      // Время бездействия пользователя в мс
}

type CreateUserRequest struct {
	Roles       []int  // Список идентификаторов ролей
	FirstName   string // Имя
	LastName    string // Фамилия
	Email       string `validate:"required"` // Электронный адрес
	Password    string `validate:"required"` // Пароль
	Description string // Описание
}

type UpdateUserRequest struct {
	Id          int64  `validate:"required"` // Идентификатор пользователя
	Roles       []int  // Список идентификаторов ролей
	FirstName   string // Имя
	LastName    string // Фамилия
	Email       string // Электронный адрес
	Description string // Описание
	Blocked     bool   // Флаг блокировки
}

type DeleteResponse struct {
	Deleted int // Количество удалённых УЗ
}

type IdentitiesRequest struct {
	Ids []int64 `validate:"required"` // Список идентификаторов
}

type IdRequest struct {
	UserId int `validate:"required"` // Идентификато пользователя
}

type ChangePasswordRequest struct {
	OldPassword string `validate:"required"` // Старый пароль
	NewPassword string `validate:"required"` // Новый пароль
}
