package routes

import (
	"github.com/integration-system/isp-kit/cluster"
	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/grpc/endpoint"
	"github.com/integration-system/isp-kit/grpc/isp"
	"msp-admin-service/controller"
)

type Controllers struct {
	Auth          controller.Auth
	User          controller.User
	Customization controller.Customization
	Secure        controller.Secure
	Session       controller.Session
	Audit         controller.Audit
}

func EndpointDescriptors() []cluster.EndpointDescriptor {
	return endpointDescriptors(Controllers{})
}

func Handler(wrapper endpoint.Wrapper, c Controllers) isp.BackendServiceServer {
	muxer := grpc.NewMux()
	for _, descriptor := range endpointDescriptors(c) {
		muxer.Handle(descriptor.Path, wrapper.Endpoint(descriptor.Handler))
	}
	return muxer
}

//nolint:funlen
func endpointDescriptors(c Controllers) []cluster.EndpointDescriptor {
	return []cluster.EndpointDescriptor{
		{
			Path:             "admin/auth/login",
			Inner:            false,
			UserAuthRequired: false,
			Handler:          c.Auth.Login,
		},
		{
			Path:             "admin/auth/login_with_sudir",
			Inner:            false,
			UserAuthRequired: false,
			Handler:          c.Auth.LoginWithSudir,
		},
		{
			Path:             "admin/auth/logout",
			Inner:            true,
			UserAuthRequired: false,
			Handler:          c.Auth.Logout,
		},
		{
			Path:             "admin/user/get_profile",
			Inner:            true,
			UserAuthRequired: false,
			Handler:          c.User.GetProfile,
		},
		{
			Path:             "admin/user/get_design",
			Inner:            true,
			UserAuthRequired: false,
			Handler:          c.Customization.GetUIDesign,
		},
		{
			Path:             "admin/user/get_users",
			Inner:            true,
			UserAuthRequired: false,
			Handler:          c.User.GetUsers,
		},
		{
			Path:             "admin/user/create_user",
			Inner:            true,
			UserAuthRequired: false,
			Handler:          c.User.CreateUser,
		},
		{
			Path:             "admin/user/update_user",
			Inner:            true,
			UserAuthRequired: false,
			Handler:          c.User.UpdateUser,
		},
		{
			Path:             "admin/user/delete_user",
			Inner:            true,
			UserAuthRequired: false,
			Handler:          c.User.DeleteUser,
		},
		{
			Path:             "admin/user/block_user",
			Inner:            true,
			UserAuthRequired: false,
			Handler:          c.User.Block,
		},
		{
			Path:             "admin/user/get_roles",
			Inner:            true,
			UserAuthRequired: false,
			Handler:          c.User.GetRoles,
		},
		{
			Path:             "admin/user/get_by_id",
			Inner:            true,
			UserAuthRequired: false,
			Handler:          c.User.GetById,
		},
		{
			Path:             "admin/session/all",
			Inner:            true,
			UserAuthRequired: false,
			Handler:          c.Session.All,
		},
		{
			Path:             "admin/session/revoke",
			Inner:            true,
			UserAuthRequired: false,
			Handler:          c.Session.Revoke,
		},
		{
			Path:             "admin/log/all",
			Inner:            true,
			UserAuthRequired: false,
			Handler:          c.Audit.All,
		},
		{
			Path:             "admin/secure/authenticate",
			Inner:            true,
			UserAuthRequired: false,
			Handler:          c.Secure.Authenticate,
		},
	}
}
