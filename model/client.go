package model

import (
	"github.com/integration-system/isp-lib/database"
	log "github.com/integration-system/isp-log"
	"github.com/integration-system/isp-log/stdcodes"
)

var (
	DbClient = database.NewRxDbClient(
		database.WithSchemaEnsuring(),
		database.WithSchemaAutoInjecting(),
		database.WithMigrationsEnsuring(),
		database.WithInitializingErrorHandler(func(err *database.ErrorEvent) {
			log.Fatal(stdcodes.InitializingDbError, err.Error())
		}),
	)
)
