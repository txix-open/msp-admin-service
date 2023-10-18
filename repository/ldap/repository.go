package ldap

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/pkg/errors"
	"msp-admin-service/conf"
)

const (
	dialTimeout = 5 * time.Second
)

type Repository struct {
	baseDn string
	conn   *ldap.Conn
}

func NewRepository(config *conf.Ldap) (*Repository, error) {
	if config == nil {
		return nil, errors.New("ldap config is not initialized")
	}

	c, err := net.DialTimeout("tcp", config.Address, dialTimeout)
	if err != nil {
		return nil, errors.WithMessagef(err, "net dial to %s", config.Address)
	}

	conn := ldap.NewConn(c, false)
	conn.Start()

	err = conn.Bind(config.Username, config.Password)
	if err != nil {
		return nil, errors.WithMessagef(err, "ldap auth, username: %s", config.Username)
	}

	return &Repository{
		conn:   conn,
		baseDn: config.BaseDn,
	}, nil
}

func (r Repository) IsExist(ctx context.Context, dn string) (bool, error) {
	request := ldap.NewSearchRequest(
		dn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=*)",
		[]string{"name"},
		nil,
	)
	result, err := r.conn.Search(request)
	if err != nil {
		return false, errors.WithMessage(err, "search by full dn")
	}
	return len(result.Entries) > 0, nil
}

func (r Repository) DnByUserPrincipalName(ctx context.Context, principalName string) (string, error) {
	filter := fmt.Sprintf("(&(userPrincipalName=%s))", principalName)
	request := ldap.NewSearchRequest(
		r.baseDn, // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"name"},
		nil,
	)
	result, err := r.conn.Search(request)
	if err != nil {
		return "", errors.WithMessagef(err, "search by '%s'", filter)
	}
	if len(result.Entries) == 0 {
		return "", errors.Errorf("not found entities by principalName: %s", principalName)
	}

	return result.Entries[0].DN, nil
}

func (r Repository) RemoveFromGroup(ctx context.Context, userDn string, groupDn string) error {
	modifyReq := ldap.NewModifyRequest(groupDn, nil)
	modifyReq.Delete("member", []string{userDn})
	err := r.conn.Modify(modifyReq)
	if err != nil {
		return errors.WithMessagef(err, "modify entity by dn %s", groupDn)
	}
	return nil
}

func (r Repository) Close() error {
	return errors.WithMessage(r.conn.Close(), "close ldap connection")
}
