package assembly

import (
	"context"

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
	"msp-admin-service/conf"
	ldapRepo "msp-admin-service/repository/ldap"
	"msp-admin-service/service/delete_old_audit_worker"
	"msp-admin-service/service/inactive_worker"
	"msp-admin-service/service/ldap"
)

type Assembly struct {
	boot     *bootstrap.Bootstrap
	db       *dbrx.Client
	server   *grpc.Server
	httpCli  *httpcli.Client
	logger   *log.Adapter
	bgjobCli *bgjobx.Client
}

func New(boot *bootstrap.Bootstrap) (*Assembly, error) {
	logger := boot.App.Logger()
	server := grpc.NewServer()
	httpCli := httpclix.Default(httpcli.WithMiddlewares(httpclix.Log(logger)))
	db := dbrx.New(logger, dbx.WithMigrationRunner(boot.MigrationsDir, logger))
	bgjobCli := bgjobx.NewClient(db, logger)
	return &Assembly{
		boot:     boot,
		db:       db,
		server:   server,
		logger:   logger,
		httpCli:  httpCli,
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

	locator := NewLocator(a.logger, a.httpCli, a.db)
	config := locator.Config(ctx, func(config *conf.Ldap) (ldap.Repo, error) {
		repo, err := ldapRepo.NewRepository(config)
		if err != nil {
			return nil, errors.WithMessage(err, "new repository")
		}
		return repo, nil
	}, newCfg)

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
