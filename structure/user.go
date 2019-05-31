package structure

import (
	"msp-admin-service/entity"
)

type UsersResponse struct {
	Items *[]entity.AdminUser `json:"items,omitempty"`
}

type UsersRequest struct {
	Ids    []int64 `json:"ids"`
	Offset int     `json:"offset"`
	Limit  int     `json:"limit"`
	Email  string  `json:"email"`
	Phone  string  `json:"phone"`
}

type AdminUserShort struct {
	TableName string `sql:"admin_service.users" json:"-"`
	Image     string `json:"image"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email" valid:"required~Required"`
	Phone     string `json:"phone"`
}
