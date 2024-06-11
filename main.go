package main

import (
	"github.com/txix-open/isp-kit/bootstrap"
	"github.com/txix-open/isp-kit/shutdown"
	"msp-admin-service/assembly"
	"msp-admin-service/conf"
	"msp-admin-service/routes"
)

var version = "1.0.0"

// @title msp-admin-service
// @version 1.0.0
// @description сервис управления администраторами

// @license.name GNU GPL v3.0

// @host localhost:9000
// @BasePath /api/admin

//go:generate swag init
//go:generate rm -f docs/swagger.json

func main() {
	boot := bootstrap.New(version, conf.Remote{}, routes.EndpointDescriptors())
	app := boot.App
	logger := app.Logger()

	assembly, err := assembly.New(boot)
	if err != nil {
		logger.Fatal(app.Context(), err)
	}
	app.AddRunners(assembly.Runners()...)
	app.AddClosers(assembly.Closers()...)

	shutdown.On(func() {
		logger.Info(app.Context(), "starting shutdown")
		app.Shutdown()
		logger.Info(app.Context(), "shutdown completed")
	})

	err = app.Run()
	if err != nil {
		app.Shutdown()
		logger.Fatal(app.Context(), err)
	}
}
