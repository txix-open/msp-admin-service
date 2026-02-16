package transaction

import (
	"context"

	"msp-admin-service/repository"
	"msp-admin-service/service"
	"msp-admin-service/service/session_worker"

	"github.com/txix-open/isp-kit/db"
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

type tokenTx struct {
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

func (m Manager) TokenTransaction(ctx context.Context, msgTx func(ctx context.Context, tx session_worker.TokenTransaction) error) error {
	return m.db.RunInTransaction(ctx, func(ctx context.Context, tx *db.Tx) error {
		token := repository.NewToken(tx)
		return msgTx(ctx, tokenTx{token})
	})
}
