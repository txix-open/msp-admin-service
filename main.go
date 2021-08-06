package main

import (
	"context"
	"net"
	"os"

	"github.com/integration-system/isp-lib/v2/backend"
	"github.com/integration-system/isp-lib/v2/bootstrap"
	"github.com/integration-system/isp-lib/v2/config/schema"
	"github.com/integration-system/isp-lib/v2/metric"
	"github.com/integration-system/isp-lib/v2/structure"
	log "github.com/integration-system/isp-log"
	"github.com/integration-system/isp-log/stdcodes"
	"github.com/soheilhy/cmux"
	"msp-admin-service/conf"
	_ "msp-admin-service/docs"
	"msp-admin-service/helper"
	"msp-admin-service/invoker"
	"msp-admin-service/model"
	"msp-admin-service/service"
)

var (
	version = "0.1.0"
	grpcLn  net.Listener
	wsLn    net.Listener
)

// @title msp-admin-service
// @version 1.0.0
// @description сервис администрирования

// @license.name GNU GPL v3.0

// @host localhost:9000
// @BasePath /api/admin

//go:generate swag init --parseDependency
//go:generate rm -f docs/swagger.json

func main() {
	bootstrap.
		ServiceBootstrap(&conf.Configuration{}, &conf.RemoteConfig{}).
		OnLocalConfigLoad(onLocalConfigLoad).
		DefaultRemoteConfigPath(schema.ResolveDefaultConfigPath("default_remote_config.json")).
		SocketConfiguration(socketConfiguration).
		DeclareMe(routesData).
		RequireModule("config", invoker.ConfigClient.ReceiveAddressList, true).
		RequireRoutes(service.SessionManager.RoutesUpdateSessionCallback).
		OnRemoteConfigReceive(onRemoteConfigReceive).
		OnShutdown(onShutdown).
		Run()
}

func socketConfiguration(cfg interface{}) structure.SocketConfiguration {
	appConfig := cfg.(*conf.Configuration)
	return structure.SocketConfiguration{
		Host:   appConfig.ConfigServiceAddress.IP,
		Port:   appConfig.ConfigServiceAddress.Port,
		Secure: false,
		UrlParams: map[string]string{
			"module_name": appConfig.ModuleName,
		},
	}
}

func onShutdown(ctx context.Context, _ os.Signal) {
	service.SessionManager.ShutdownSocket(ctx)
	backend.StopGrpcServer()
	closeListeners(wsLn, grpcLn)
}

func onRemoteConfigReceive(remoteConfig, oldRemoteConfig *conf.RemoteConfig) {
	model.DbClient.ReceiveConfiguration(remoteConfig.Database)
	metric.InitCollectors(remoteConfig.Metrics, oldRemoteConfig.Metrics)
	metric.InitHttpServer(remoteConfig.Metrics)
}

func onLocalConfigLoad(cfg *conf.Configuration) {
	handlers := helper.GetHandlers()
	defaultService := backend.GetDefaultService(cfg.ModuleName, handlers...)

	if err := serve(cfg.GrpcInnerAddress, defaultService); err != nil {
		log.Fatal(stdcodes.ModuleInvalidLocalConfig, err)
	}
}

func serve(addr structure.AddressConfiguration, grpcService *backend.DefaultService) error {
	ln, err := net.Listen("tcp", addr.GetAddress())
	if err != nil {
		return err
	}

	m := cmux.New(ln)
	grpcLn = m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	wsLn = m.Match(cmux.HTTP1HeaderField("Upgrade", "websocket"))

	service.SessionManager.InitWebSocket(wsLn)

	backend.StartBackendGrpcServerOn(addr, grpcLn, grpcService)
	go func() {
		if err := m.Serve(); err != nil {
			log.Warn(stdcodes.ModuleGrpcServiceStartError, err)
		}
	}()

	return nil
}

func routesData(localConfig interface{}) bootstrap.ModuleInfo {
	cfg := localConfig.(*conf.Configuration)
	return bootstrap.ModuleInfo{
		ModuleName:       cfg.ModuleName,
		ModuleVersion:    version,
		GrpcOuterAddress: cfg.GrpcOuterAddress,
		Handlers:         helper.GetHandlers(),
	}
}

func closeListeners(lns ...net.Listener) {
	for _, ln := range lns {
		if ln != nil {
			_ = ln.Close()
		}
	}
}
