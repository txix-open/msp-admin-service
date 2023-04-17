package assembly

import (
	"context"
	"time"

	"github.com/integration-system/isp-kit/app"
	"github.com/integration-system/isp-kit/bootstrap"
	"github.com/integration-system/isp-kit/cluster"
	"github.com/integration-system/isp-kit/dbrx"
	"github.com/integration-system/isp-kit/dbx"
	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/http/httpcli"
	"github.com/integration-system/isp-kit/http/httpclix"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/worker"
	"github.com/pkg/errors"
	"msp-admin-service/conf"
)

type Assembly struct {
	boot    *bootstrap.Bootstrap
	db      *dbrx.Client
	server  *grpc.Server
	httpCli *httpcli.Client
	worker  *worker.Worker
	logger  *log.Adapter
}

func New(boot *bootstrap.Bootstrap) (*Assembly, error) {
	server := grpc.NewServer()
	httpCli := httpclix.Default(httpcli.WithMiddlewares(httpclix.Log(boot.App.Logger())))
	db := dbrx.New(dbx.WithMigration(boot.MigrationsDir))
	return &Assembly{
		boot:    boot,
		db:      db,
		server:  server,
		logger:  boot.App.Logger(),
		httpCli: httpCli,
	}, nil
}

func (a *Assembly) ReceiveConfig(ctx context.Context, remoteConfig []byte) error {
	var (
		newCfg  conf.Remote
		prevCfg conf.Remote
	)
	err := a.boot.RemoteConfig.Upgrade(remoteConfig, &newCfg, &prevCfg)
	if err != nil {
		a.logger.Fatal(ctx, errors.WithMessage(err, "upgrade remote config"))
	}

	a.logger.SetLevel(newCfg.LogLevel)

	err = a.db.Upgrade(ctx, newCfg.Database)
	if err != nil {
		a.logger.Fatal(ctx, errors.WithMessage(err, "upgrade db client"), log.Any("config", a.hiddenSecret(newCfg.Database)))
	}

	locator := NewLocator(a.logger, a.httpCli, a.db)
	config := locator.Config(newCfg)

	a.server.Upgrade(config.Handler)

	if a.worker != nil {
		a.worker.Shutdown()
	}
	a.worker = worker.New(
		config.InactiveBlocker,
		worker.WithInterval(time.Duration(newCfg.BlockInactiveWorker.RunIntervalInMinutes)*time.Minute),
	)
	a.worker.Run(a.boot.App.Context())

	return nil
}

func (a *Assembly) Runners() []app.Runner {
	eventHandler := cluster.NewEventHandler().
		RemoteConfigReceiver(a)
	return []app.Runner{
		app.RunnerFunc(func(ctx context.Context) error {
			return a.server.ListenAndServe(a.boot.BindingAddress)
		}),
		app.RunnerFunc(func(ctx context.Context) error {
			return a.boot.ClusterCli.Run(ctx, eventHandler)
		}),
	}
}

func (a *Assembly) Closers() []app.Closer {
	return []app.Closer{
		a.boot.ClusterCli,
		app.CloserFunc(func() error {
			a.server.Shutdown()
			return nil
		}),
		a.db,
	}
}

func (a *Assembly) hiddenSecret(conf dbx.Config) dbx.Config {
	conf.Password = "***"
	return conf
}
