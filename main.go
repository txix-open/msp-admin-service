package main

import (
	"admin-service/conf"
	"admin-service/helper"
	"context"
	"gitlab.alx/msp2.0/msp-lib/backend"
	"gitlab.alx/msp2.0/msp-lib/bootstrap"
	"gitlab.alx/msp2.0/msp-lib/database"
	"gitlab.alx/msp2.0/msp-lib/socket"
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
		SocketConfiguration(socketConfiguration).
		SendRemoteConfigSchemaWithVersion(version).
		SendRoutes(routesData).
		OnRemoteConfigReceive(onRemoteConfigReceive).
		OnShutdown(onShutdown).
		Run()
}

func socketConfiguration(cfg interface{}) socket.SocketConfiguration {
	appConfig := cfg.(*conf.Configuration)
	return socket.SocketConfiguration{
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

func onRemoteConfigReceive(remoteConfig, _ *conf.RemoteConfig) {
	database.InitDb(remoteConfig.Database)
}

func onLocalConfigLoad(cfg *conf.Configuration) {
	handlers := helper.GetHandlers()
	service := backend.GetDefaultService(cfg.ModuleName, handlers...)
	backend.StartBackendGrpcServer(cfg.GrpcInnerAddress, service)
}

func routesData(localConfig interface{}) bootstrap.SendRoutesData {
	cfg := localConfig.(*conf.Configuration)
	return bootstrap.SendRoutesData{
		ModuleName:       cfg.ModuleName,
		Version:          version,
		GrpcOuterAddress: cfg.GrpcOuterAddress,
		Handlers:         helper.GetHandlers(),
	}
}
