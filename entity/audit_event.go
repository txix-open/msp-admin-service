package entity

const (
	EventSuccessLogin        = "success_login"
	EventErrorLogin          = "error_login"
	EventSuccessLogout       = "success_logout"
	EventRoleChanged         = "role_changed"
	EventUserChanged         = "user_changed"
	EventUserPasswordChanged = "success_change_password"
	EventErrorPasswordChange = "error_change_password"
	EventUserBlocked         = "user_blocked"
)

type AuditEvent struct {
	Event  string
	Enable bool
}
