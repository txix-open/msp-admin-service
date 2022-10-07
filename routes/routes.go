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
			Path:             "admin/secure/authenticate",
			Inner:            true,
			UserAuthRequired: false,
			Handler:          c.Secure.Authenticate,
		},
	}
}
