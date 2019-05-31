package helper

import (
	"google.golang.org/grpc/metadata"
	"msp-admin-service/controller"
	"msp-admin-service/entity"
	"msp-admin-service/structure"
)

type Handlers struct {
	// ===== AUTH =====
	Login  func(structure.AuthRequest) (*structure.Auth, error) `method:"login" group:"auth" inner:"false"`
	Logout func(metadata.MD) error                              `method:"logout" group:"auth" inner:"true"`
	// ===== USER =====
	GetProfile       func(metadata.MD) (*structure.AdminUserShort, error)                            `method:"get_profile" group:"user" inner:"true"`
	GetUsers         func(structure.UsersRequest) (*structure.UsersResponse, error)                  `method:"get_users" group:"user" inner:"true"`
	CreateUpdateUser func(user entity.AdminUser) (*entity.AdminUser, error)                          `method:"create_update_user" group:"user" inner:"true"`
	DeleteUser       func(identities structure.IdentitiesRequest) (*structure.DeleteResponse, error) `method:"delete_user" group:"user" inner:"true"`
}

func GetHandlers() []interface{} {
	return []interface{}{
		&Handlers{
			Login:            controller.Login,
			Logout:           controller.Logout,
			GetProfile:       controller.GetProfile,
			GetUsers:         controller.GetUsers,
			CreateUpdateUser: controller.CreateUpdateUser,
			DeleteUser:       controller.DeleteUser,
		},
	}
}
