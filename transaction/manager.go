package transaction

import (
	"context"

	"github.com/txix-open/isp-kit/db"
	"msp-admin-service/repository"
	"msp-admin-service/service"
)

type Manager struct {
	db db.Transactional
}

func NewManager(db db.Transactional) *Manager {
	return &Manager{
		db: db,
	}
}

type userTx struct {
	repository.User
	repository.Role
	repository.UserRole
	repository.Token
}

func (m Manager) UserTransaction(ctx context.Context, msgTx func(ctx context.Context, tx service.UserTransaction) error) error {
	return m.db.RunInTransaction(ctx, func(ctx context.Context, tx *db.Tx) error {
		user := repository.NewUser(tx)
		role := repository.NewRole(tx)
		userRole := repository.NewUserRole(tx)
		token := repository.NewToken(tx)
		return msgTx(ctx, userTx{user, role, userRole, token})
	})
}

func (m Manager) AuthTransaction(ctx context.Context, msgTx func(ctx context.Context, tx service.AuthTransaction) error) error {
	return m.db.RunInTransaction(ctx, func(ctx context.Context, tx *db.Tx) error {
		user := repository.NewUser(tx)
		role := repository.NewRole(tx)
		userRole := repository.NewUserRole(tx)
		token := repository.NewToken(tx)
		return msgTx(ctx, userTx{user, role, userRole, token})
	})
}
