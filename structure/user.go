package structure

import "time"

type User struct {
	TableName string    `sql:"admin_service.users" json:"-"`
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

type UsersResponse struct {
	Items *[]User `json:"items,omitempty"`
}

type UsersRequest struct {
	Ids    []int64 `json:"ids"`
	Offset int     `json:"offset"`
	Limit  int     `json:"limit"`
	Email  string  `json:"email"`
	Phone  string  `json:"phone"`
}
