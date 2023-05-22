package service

import (
	"context"
	"fmt"

	"github.com/integration-system/isp-kit/json"
	"msp-admin-service/conf"
	"msp-admin-service/domain"
)

type Permission struct {
	permissions []domain.Permission
}

func NewPermission(permissions []conf.Permission) *Permission {
	perms := []string{}

	for _, perm := range permissions {
		perms = append(perms, perm.Key)
	}

	bb, _ := json.Marshal(perms)
	fmt.Println(string(bb))
	return &Permission{
		permissions: toDomain(permissions),
	}
}
func (s *Permission) All(_ context.Context) []domain.Permission {
	return s.permissions
}

func toDomain(permissions []conf.Permission) []domain.Permission {
	permList := make([]domain.Permission, 0)
	for _, perm := range permissions {
		permList = append(permList, domain.Permission{
			Key:  perm.Key,
			Name: perm.Name,
		})
	}

	return permList
}
