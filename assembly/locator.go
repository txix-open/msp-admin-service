package assembly

import (
	"context"
	"time"

	"github.com/integration-system/isp-kit/bgjobx"
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
	"msp-admin-service/service/delete_old_audit_worker"
	"msp-admin-service/service/inactive_worker"
	"msp-admin-service/service/ldap"
	"msp-admin-service/transaction"
)

var (
	JobPollInterval = 1 * time.Minute //nolint:gochecknoglobals
)

type DB interface {
	db.DB
	db.Transactional
}

type Locator struct {
	logger  log.Logger
	httpCli *httpcli.Client
	db      DB
}

func NewLocator(logger log.Logger, httpCli *httpcli.Client, db DB) Locator {
	return Locator{
		logger:  logger,
		httpCli: httpCli,
		db:      db,
	}
}

type Config struct {
	Handler  isp.BackendServiceServer
	BgJobCfg []bgjobx.WorkerConfig
}

func (l Locator) Config(ctx context.Context, ldapRepoSupplier ldap.RepoSupplier, cfg conf.Remote) Config {
	sudirRepo := repository.NewSudir(l.httpCli, cfg.SudirAuth)
	roleRepo := repository.NewRole(l.db)
	userRepo := repository.NewUser(l.db)
	tokenRepo := repository.NewToken(l.db)
	auditRepo := repository.NewAudit(l.db)
	auditEventRepo := repository.NewAuditEvent(l.db)
	userRoleRepo := repository.NewUserRole(l.db)

	auditService := service.NewAudit(ctx, l.logger, auditRepo, auditEventRepo, cfg.Audit.EventSettings)
	tokenService := service.NewToken(tokenRepo, cfg.ExpireSec)
	sudirService := service.NewSudir(cfg.SudirAuth, sudirRepo)

	txManager := transaction.NewManager(l.db)

	userService := service.NewUser(userRepo, userRoleRepo, roleRepo, tokenRepo, auditService, txManager, l.logger)
	authService := service.NewAuth(
		userRepo, txManager, tokenService, sudirService, auditService, l.logger,
		cfg.AntiBruteforce.DelayLoginRequestInSec,
		cfg.AntiBruteforce.MaxInFlightLoginRequests,
	)
	roleService := service.NewRole(roleRepo, auditService)

	permissionsService := service.NewPermission(cfg.Permissions)

	userController := controller.NewUser(userService)
	customizationController := controller.NewCustomization(cfg.UiDesign)
	authController := controller.NewAuth(authService, l.logger)
	secureController := controller.NewSecure(tokenService)
	sessionController := controller.NewSession(tokenService)
	auditController := controller.NewAudit(auditService)
	roleController := controller.NewRole(roleService)
	permissionController := controller.NewPermissions(permissionsService)

	handler := routes.Handler(
		endpoint.DefaultWrapper(l.logger),
		routes.Controllers{
			User:          userController,
			Customization: customizationController,
			Auth:          authController,
			Secure:        secureController,
			Session:       sessionController,
			Audit:         auditController,
			Role:          roleController,
			Permissions:   permissionController,
		},
	)

	ldapService := ldap.NewService(cfg.Ldap, ldapRepoSupplier, userRoleRepo, roleRepo, l.logger)
	inactiveBlocker := inactive_worker.NewInactiveBlocker(
		tokenRepo, userRepo, auditService, ldapService,
		cfg.BlockInactiveWorker,
		l.logger,
	)

	deleteOldAuditWorker := delete_old_audit_worker.NewService(l.logger, auditRepo, cfg.Audit.AuditTTl)

	return Config{
		Handler: handler,
		BgJobCfg: []bgjobx.WorkerConfig{{
			Queue:        delete_old_audit_worker.QueueName,
			Concurrency:  1,
			PollInterval: JobPollInterval,
			Handle:       deleteOldAuditWorker,
		}, {
			Queue:        inactive_worker.QueueName,
			Concurrency:  1,
			PollInterval: JobPollInterval,
			Handle:       inactiveBlocker,
		}},
	}
}
