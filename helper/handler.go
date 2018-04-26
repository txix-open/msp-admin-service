package helper

import (
	"admin-service/structure"
	"admin-service/controller"
)

type Handlers struct {
	// ===== AUTH =====
	Auth func() (*structure.Auth, error) `method:"auth"`
}

func GetHandlers() *Handlers {
	return &Handlers{
		controller.Auth,
	}
}
