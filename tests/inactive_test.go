package tests_test

import (
	"msp-admin-service/service/inactive_worker"
	"testing"
	"time"

	"github.com/txix-open/isp-kit/bgjobx"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
	"msp-admin-service/assembly"
	"msp-admin-service/conf"
	"msp-admin-service/entity"
	"msp-admin-service/repository"
)

func TestInactiveWorker(t *testing.T) {
	t.Parallel()

	test, require := test.New(t)
	db := dbt.New(test, dbx.WithMigrationRunner("../migrations", test.Logger()))

	userId := InsertUser(db, entity.User{Email: "a@test", LastActiveAt: time.Now().UTC().Add(-5 * 24 * time.Hour)})
	InsertUser(db, entity.User{Email: "b@test", LastActiveAt: time.Now().UTC()})

	config := assembly.NewLocator(test.Logger(), nil, db).
		Config(t.Context(), emptyLdap, conf.Remote{
			BlockInactiveWorker: conf.BlockInactiveWorker{
				DaysThreshold:        1,
				RunIntervalInMinutes: 1,
			},
		}, 500*time.Millisecond)

	bgjobCli := bgjobx.NewClient(db, test.Logger())
	err := inactive_worker.EnqueueSeedJob(t.Context(), bgjobCli)
	require.NoError(err)
	err = bgjobCli.Upgrade(t.Context(), config.BgJobCfg)
	require.NoError(err)

	time.Sleep(5 * time.Second)

	user, err := repository.NewUser(db).GetUserById(t.Context(), userId)
	require.NoError(err)
	require.True(user.Blocked)

	list, err := repository.NewAudit(db).All(t.Context(), 10, 0)
	require.NoError(err)
	require.Len(list, 1)
}
