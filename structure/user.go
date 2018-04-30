package structure

import (
	"gitlab8.alx/msp2.0/msp-lib/structure"
)

type UsersResponse struct {
	Items *[]structure.AdminUser `json:"items,omitempty"`
}

type UsersRequest struct {
	Ids    []int64 `json:"ids"`
	Offset int     `json:"offset"`
	Limit  int     `json:"limit"`
	Email  string  `json:"email"`
	Phone  string  `json:"phone"`
}
