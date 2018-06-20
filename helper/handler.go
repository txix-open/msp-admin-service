package helper

import (
	libStr "gitlab8.alx/msp2.0/msp-lib/structure"
	"admin-service/structure"
	"admin-service/controller"
	"google.golang.org/grpc/metadata"
)

type Handlers struct {
	// ===== AUTH =====
	Login            func(structure.AuthRequest) (*structure.Auth, error)                            `method:"login" group:"auth" inner:"false"`
	Logout           func(metadata.MD) error                                                         `method:"logout" group:"auth" inner:"true"`
	// ===== USER =====
	GetProfile       func(metadata.MD) (*structure.AdminUserShort, error)                            `method:"get_profile" group:"user" inner:"true"`
	GetUsers         func(structure.UsersRequest) (*structure.UsersResponse, error)                  `method:"get_users" group:"user" inner:"true"`
	CreateUpdateUser func(user libStr.AdminUser) (*libStr.AdminUser, error)                          `method:"create_update_user" group:"user" inner:"true"`
	DeleteUser       func(identities structure.IdentitiesRequest) (*structure.DeleteResponse, error) `method:"delete_user" group:"user" inner:"true"`
}

func GetHandlers() *Handlers {
	return &Handlers{
		controller.Login,
		controller.Logout,
		controller.GetProfile,
		controller.GetUsers,
		controller.CreateUpdateUser,
		controller.DeleteUser,
	}
}
