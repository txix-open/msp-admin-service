package assembly

import (
	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/grpc/endpoint"
	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/integration-system/isp-kit/http/httpcli"
	"github.com/integration-system/isp-kit/log"
	"msp-admin-service/conf"
	"msp-admin-service/controller"
	"msp-admin-service/repository"
	"msp-admin-service/routes"
	"msp-admin-service/service"
	"msp-admin-service/service/worker"
)

type Locator struct {
	logger  log.Logger
	httpCli *httpcli.Client
	db      db.DB
}

func NewLocator(logger log.Logger, httpCli *httpcli.Client, db db.DB) Locator {
	return Locator{
		logger:  logger,
		httpCli: httpCli,
		db:      db,
	}
}

type Config struct {
	Handler         isp.BackendServiceServer
	InactiveBlocker worker.InactiveBlocker
}

func (l Locator) Handler(cfg conf.Remote) isp.BackendServiceServer {
	return l.Config(cfg).Handler
}

func (l Locator) Config(cfg conf.Remote) Config {
	sudirRepo := repository.NewSudir(l.httpCli, cfg.SudirAuth)
	roleRepo := repository.NewRole(l.db)
	userRepo := repository.NewUser(l.db)
	tokenRepo := repository.NewToken(l.db)
	auditRepo := repository.NewAudit(l.db)

	auditService := service.NewAudit(auditRepo, l.logger)
	tokenService := service.NewToken(tokenRepo, cfg.ExpireSec)
	sudirService := service.NewSudir(cfg.SudirAuth, sudirRepo, roleRepo)
	userService := service.NewUser(userRepo, roleRepo, tokenRepo, l.logger)
	authService := service.NewAuth(
		userRepo,
		tokenService,
		sudirService,
		auditService,
		l.logger,
		cfg.AntiBruteforce.DelayLoginRequestInSec,
		cfg.AntiBruteforce.MaxInFlightLoginRequests,
	)

	userController := controller.NewUser(userService)
	customizationController := controller.NewCustomization(cfg.UiDesign)
	authController := controller.NewAuth(authService, l.logger)
	secureController := controller.NewSecure(tokenService)
	sessionController := controller.NewSession(tokenService)
	auditController := controller.NewAudit(auditService)

	handler := routes.Handler(
		endpoint.DefaultWrapper(l.logger),
		routes.Controllers{
			User:          userController,
			Customization: customizationController,
			Auth:          authController,
			Secure:        secureController,
			Session:       sessionController,
			Audit:         auditController,
		},
	)

	inactiveBlocker := worker.NewInactiveBlocker(
		tokenRepo,
		userRepo,
		auditService,
		cfg.BlockInactiveWorker.DaysThreshold,
		l.logger,
	)

	return Config{
		Handler:         handler,
		InactiveBlocker: inactiveBlocker,
	}
}
