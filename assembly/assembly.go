package assembly

import (
	"context"
	"time"

	"msp-admin-service/conf"
	"msp-admin-service/service/delete_old_audit_worker"
	"msp-admin-service/service/inactive_worker"
	"msp-admin-service/service/session_worker"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/app"
	"github.com/txix-open/isp-kit/bgjobx"
	"github.com/txix-open/isp-kit/bootstrap"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/dbrx"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/http/httpclix"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/observability/sentry"
)

type Assembly struct {
	boot     *bootstrap.Bootstrap
	db       *dbrx.Client
	server   *grpc.Server
	sudirCli *httpcli.Client
	logger   *log.Adapter
	bgjobCli *bgjobx.Client
}

func New(boot *bootstrap.Bootstrap) (*Assembly, error) {
	logger := boot.App.Logger()
	wrappedLogger := sentry.WrapErrorLogger(logger, boot.SentryHub)

	db := dbrx.New(logger, dbx.WithMigrationRunner(boot.MigrationsDir, wrappedLogger))
	boot.HealthcheckRegistry.Register("db", db)

	bgjobCli := bgjobx.NewClient(db, wrappedLogger)
	return &Assembly{
		boot:     boot,
		db:       db,
		server:   grpc.NewServer(),
		logger:   logger,
		sudirCli: httpclix.Default(httpcli.WithMiddlewares(httpclix.Log(wrappedLogger))),
		bgjobCli: bgjobCli,
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
		a.logger.Fatal(ctx, errors.WithMessage(err, "upgrade db client"))
	}

	sudirBaseUrl := ""
	if newCfg.SudirAuth != nil {
		sudirBaseUrl = newCfg.SudirAuth.Host
	}
	a.sudirCli.GlobalRequestConfig().BaseUrl = sudirBaseUrl

	locator := NewLocator(a.logger, a.sudirCli, a.db)
	config, err := locator.Config(ctx, newCfg, time.Minute)
	if err != nil {
		a.logger.Fatal(ctx, errors.WithMessage(err, "locator config"))
	}

	a.server.Upgrade(config.Handler)

	err = a.bgjobCli.Upgrade(a.boot.App.Context(), config.BgJobCfg)
	if err != nil {
		a.logger.Fatal(ctx, errors.WithMessage(err, "upgrade bgjob client"))
	}

	err = delete_old_audit_worker.EnqueueSeedJob(ctx, a.bgjobCli)
	if err != nil {
		a.logger.Fatal(ctx, errors.WithMessage(err, "seed delete old audit worker"))
	}
	err = inactive_worker.EnqueueSeedJob(ctx, a.bgjobCli)
	if err != nil {
		a.logger.Fatal(ctx, errors.WithMessage(err, "seed inactive user worker"))
	}
	err = session_worker.EnqueueSeedJob(ctx, a.bgjobCli)
	if err != nil {
		a.logger.Fatal(ctx, errors.WithMessage(err, "seed expire session worker"))
	}

	return nil
}

func (a *Assembly) Runners() []app.Runner {
	eventHandler := cluster.NewEventHandler().
		RemoteConfigReceiver(a)
	return []app.Runner{
		app.RunnerFunc(func(ctx context.Context) error {
			err := a.server.ListenAndServe(a.boot.BindingAddress)
			if err != nil {
				return errors.WithMessage(err, "listen ans serve grpc server")
			}
			return nil
		}),
		app.RunnerFunc(func(ctx context.Context) error {
			err := a.boot.ClusterCli.Run(ctx, eventHandler)
			if err != nil {
				return errors.WithMessage(err, "run cluster client")
			}
			return nil
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
		app.CloserFunc(func() error {
			a.bgjobCli.Close()
			return nil
		}),
		a.db,
	}
}
