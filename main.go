package main

import (
	"context"
	"github.com/integration-system/isp-lib/backend"
	"github.com/integration-system/isp-lib/bootstrap"
	"github.com/integration-system/isp-lib/config/schema"
	"github.com/integration-system/isp-lib/database"
	"github.com/integration-system/isp-lib/metric"
	"github.com/integration-system/isp-lib/structure"
	"msp-admin-service/conf"
	"msp-admin-service/helper"
	"os"
)

var (
	version = "0.1.0"
	date    = "undefined"
)

func main() {
	bootstrap.
		ServiceBootstrap(&conf.Configuration{}, &conf.RemoteConfig{}).
		OnLocalConfigLoad(onLocalConfigLoad).
		DefaultRemoteConfigPath(schema.ResolveDefaultConfigPath("default_remote_config.json")).
		SocketConfiguration(socketConfiguration).
		DeclareMe(routesData).
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
			"module_name":   appConfig.ModuleName,
			"instance_uuid": appConfig.InstanceUuid,
		},
	}
}

func onShutdown(_ context.Context, _ os.Signal) {
	backend.StopGrpcServer()
}

func onRemoteConfigReceive(remoteConfig, oldRemoteConfig *conf.RemoteConfig) {
	database.InitDb(remoteConfig.Database)
	metric.InitCollectors(remoteConfig.Metrics, oldRemoteConfig.Metrics)
	metric.InitHttpServer(remoteConfig.Metrics)
}

func onLocalConfigLoad(cfg *conf.Configuration) {
	handlers := helper.GetHandlers()
	service := backend.GetDefaultService(cfg.ModuleName, handlers...)
	backend.StartBackendGrpcServer(cfg.GrpcInnerAddress, service)
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
