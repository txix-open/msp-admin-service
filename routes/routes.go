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
	Role          controller.Role
	Permissions   controller.Permissions
}

func EndpointDescriptors() []cluster.EndpointDescriptor {
	return endpointDescriptors(Controllers{})
}

func Handler(wrapper endpoint.Wrapper, c Controllers) isp.BackendServiceServer { // nolint:ireturn
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
			Path:    "admin/auth/login",
			Inner:   false,
			Handler: c.Auth.Login,
		},
		{
			Path:    "admin/auth/login_with_sudir",
			Inner:   false,
			Handler: c.Auth.LoginWithSudir,
		},
		{
			Path:    "admin/auth/logout",
			Inner:   true,
			Handler: c.Auth.Logout,
		},
		{
			Path:    "admin/user/get_profile",
			Inner:   true,
			Handler: c.User.GetProfile,
		},
		{
			Path:    "admin/user/get_design",
			Inner:   true,
			Handler: c.Customization.GetUIDesign,
		},
		{
			Path:    "admin/user/get_users",
			Inner:   true,
			Handler: c.User.GetUsers,
		},
		{
			Path:    "admin/user/create_user",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("user_create"),
			Handler: c.User.CreateUser,
		},
		{
			Path:    "admin/user/update_user",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("user_update"),
			Handler: c.User.UpdateUser,
		},
		{
			Path:    "admin/user/delete_user",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("user_delete"),
			Handler: c.User.DeleteUser,
		},
		{
			Path:    "admin/user/block_user",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("user_block"),
			Handler: c.User.Block,
		},
		{
			Path:    "admin/user/get_by_id",
			Inner:   true,
			Handler: c.User.GetById,
		},
		{
			Path:    "admin/role/all",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("role_view"),
			Handler: c.Role.All,
		},
		{
			Path:    "admin/user/get_permissions",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("role_view"),
			Handler: c.Permissions.GetPermissions,
		},
		{
			Path:    "admin/role/create",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("role_add"),
			Handler: c.Role.CreateRole,
		},
		{
			Path:    "admin/role/update",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("role_update"),
			Handler: c.Role.UpdateRole,
		},
		{
			Path:    "admin/role/delete",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("role_delete"),
			Handler: c.Role.DeleteRole,
		},
		{
			Path:    "admin/session/all",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("session_view"),
			Handler: c.Session.All,
		},
		{
			Path:    "admin/session/revoke",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("session_revoke"),
			Handler: c.Session.Revoke,
		},
		{
			Path:    "admin/log/all",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("security_log_view"),
			Handler: c.Audit.All,
		},
		{
			Path:    "admin/log/events",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("audit_management_view"),
			Handler: c.Audit.Events,
		},
		{
			Path:    "admin/log/set_events",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("audit_management_view"),
			Handler: c.Audit.SetEvents,
		},
		{
			Path:    "admin/secure/authenticate",
			Inner:   true,
			Handler: c.Secure.Authenticate,
		},
		{
			Path:    "admin/secure/authorize",
			Inner:   true,
			Handler: c.Secure.Authorize,
		},
	}
}
