package domain

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	ErrNotFound             = errors.New("not found")
	ErrUnauthenticated      = errors.New("authentication failure")
	ErrSudirAuthorization   = errors.New("User is authorized only with SUDIR")
	ErrSudirAuthIsMissed    = errors.New("SUDIR authorization is not configured on the server")
	ErrInvalid              = errors.New("entity is invalid")
	ErrAlreadyExists        = errors.New("already exists")
	ErrTokenExpired         = errors.New("token expired")
	ErrTokenNotFound        = errors.New("token not found")
	ErrTooManyLoginRequests = errors.New("too many login requests")
	ErrUserIsBlocked        = errors.New("user is blocked")
	ErrNoActionRequired     = errors.New("no action required")
)

type UnknownAuditEventError struct {
	Event string
}

func (e UnknownAuditEventError) Error() string {
	return fmt.Sprintf("unknown audit event: %s", e.Event)
}
