package helper

import (
	"admin-service/structure"
	"admin-service/controller"
)

type Handlers struct {
	// ===== AUTH =====
	Auth             func(structure.AuthRequest) (*structure.Auth, error)                            `method:"auth"`
	GetUsers         func(structure.UsersRequest) (*structure.UsersResponse, error)                  `method:"get_users"`
	CreateUpdateUser func(user structure.User) (*structure.User, error)                              `method:"create_update_user"`
	DeleteUser       func(identities structure.IdentitiesRequest) (*structure.DeleteResponse, error) `method:"delete_user"`
}

func GetHandlers() *Handlers {
	return &Handlers{
		controller.Auth,
		controller.GetUsers,
		controller.CreateUpdateUser,
		controller.DeleteUser,
	}
}
