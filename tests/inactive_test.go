package tests_test

import (
	"context"
	"testing"
	"time"

	"github.com/integration-system/isp-kit/dbx"
	"github.com/integration-system/isp-kit/test"
	"github.com/integration-system/isp-kit/test/dbt"
	"msp-admin-service/assembly"
	"msp-admin-service/conf"
	"msp-admin-service/entity"
	"msp-admin-service/repository"
	"msp-admin-service/tests"
)

func TestInactiveWorker(t *testing.T) {
	t.Parallel()

	test, require := test.New(t)
	db := dbt.New(test, dbx.WithMigration("../migrations"))

	userId := tests.InsertUser(db, entity.User{Email: "a@test"})
	tests.InsertUser(db, entity.User{Email: "b@test"})
	tests.InsertTokenEntity(db, entity.Token{
		Id:        0,
		Token:     "123",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		CreatedAt: time.Now().UTC().Add(-48 * time.Hour),
	})
	tests.InsertTokenEntity(db, entity.Token{
		Id:        0,
		Token:     "234",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		CreatedAt: time.Now().UTC().Add(-5 * 24 * time.Hour),
	})

	worker := assembly.NewLocator(test.Logger(), nil, db).
		Config(context.Background(), conf.Remote{BlockInactiveWorker: conf.BlockInactiveWorker{DaysThreshold: 1}}).
		InactiveBlocker
	worker.Do(context.Background())
	time.Sleep(1 * time.Second)

	user, err := repository.NewUser(db).GetUserById(context.Background(), userId)
	require.NoError(err)
	require.True(user.Blocked)

	list, err := repository.NewAudit(db).All(context.Background(), 10, 0)
	require.NoError(err)
	require.Len(list, 1)
}
