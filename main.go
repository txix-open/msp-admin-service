package main

import (
	"context"
	"os"

	"github.com/integration-system/isp-lib/v2/backend"
	"github.com/integration-system/isp-lib/v2/bootstrap"
	"github.com/integration-system/isp-lib/v2/config/schema"
	"github.com/integration-system/isp-lib/v2/metric"
	"github.com/integration-system/isp-lib/v2/structure"
	"msp-admin-service/conf"
	"msp-admin-service/helper"
	"msp-admin-service/model"
)

var (
	version = "0.1.0"
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
	model.DbClient.ReceiveConfiguration(remoteConfig.Database)
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
