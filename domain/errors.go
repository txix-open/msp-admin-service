package domain

import (
	"github.com/pkg/errors"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrUnauthenticated    = errors.New("authentication failure")
	ErrSudirAuthorization = errors.New("User is authorized only with SUDIR")
	ErrSudirAuthIsMissed  = errors.New("SUDIR authorization is not configured on the server")
	ErrInvalid            = errors.New("entity is invalid")
	ErrAlreadyExists      = errors.New("already exists")
)
