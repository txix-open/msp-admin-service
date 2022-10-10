package assembly

import (
	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/http/httpcli"
	"github.com/integration-system/isp-kit/log"
	"msp-admin-service/conf"
	"msp-admin-service/controller"
	"msp-admin-service/repository"
	"msp-admin-service/service"

	"github.com/integration-system/isp-kit/grpc/endpoint"
	"github.com/integration-system/isp-kit/grpc/isp"
	"msp-admin-service/routes"
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

func (l Locator) Handler(cfg conf.Remote) isp.BackendServiceServer {
	sudirRepo := repository.NewSudir(l.httpCli, cfg.SudirAuth)
	roleRepo := repository.NewRole(l.db)
	userRepo := repository.NewUser(l.db)
	tokenRepo := repository.NewToken(l.db)

	tokenService := service.NewToken(tokenRepo, cfg.ExpireSec)
	sudirService := service.NewSudir(cfg.SudirAuth, sudirRepo, roleRepo)
	userService := service.NewUser(userRepo, roleRepo, l.logger)
	authService := service.NewAuth(userRepo, tokenService, sudirService, l.logger)

	userController := controller.NewUser(userService)
	customizationController := controller.NewCustomization(cfg.UiDesign)
	authController := controller.NewAuth(authService, l.logger)
	secureController := controller.NewSecure(tokenService)

	handler := routes.Handler(
		endpoint.DefaultWrapper(l.logger),
		routes.Controllers{
			User:          userController,
			Customization: customizationController,
			Auth:          authController,
			Secure:        secureController,
		},
	)

	return handler
}
