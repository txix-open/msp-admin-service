package entity

const (
	EventSuccessLogin  = "success_login"
	EventErrorLogin    = "error_login"
	EventSuccessLogout = "success_logout"
	EventRoleChanged   = "role_changed"
	EventUserChanged   = "user_changed"
)

type AuditEvent struct {
	Event  string
	Enable bool
}
