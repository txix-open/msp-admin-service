package tests_test

import (
	"context"

	"msp-admin-service/conf"
	"msp-admin-service/service/ldap"
)

type mockLdap struct {
}

func (m mockLdap) IsExist(ctx context.Context, dn string) (bool, error) {
	return false, nil
}

func (m mockLdap) DnByUserPrincipalName(ctx context.Context, principalName string) (string, error) {
	return principalName, nil
}

func (m mockLdap) ModifyMemberAttr(ctx context.Context, userDn string, groupDn string, operation string) error {
	return nil
}

func (m mockLdap) Close() error {
	return nil
}

var (
	//nolint:gochecknoglobals
	emptyLdap = ldap.RepoSupplier(func(config *conf.Ldap) (ldap.Repo, error) {
		return mockLdap{}, nil
	})
)
