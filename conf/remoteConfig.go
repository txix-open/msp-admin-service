package conf

import (
	"gitlab.alx/msp2.0/msp-lib/database"
	"gitlab.alx/msp2.0/msp-lib/structure"
)

type RemoteConfig struct {
	Database database.DBConfiguration `valid:"required~Required" json:"database"`
	Metrics  structure.MetricConfiguration
}
