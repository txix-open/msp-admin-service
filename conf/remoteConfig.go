package conf

import (
	"github.com/integration-system/isp-lib/database"
	"github.com/integration-system/isp-lib/structure"
)

type RemoteConfig struct {
	Database database.DBConfiguration      `valid:"required~Required" json:"database" schema:"Database"`
	Metrics  structure.MetricConfiguration `schema:"Metrics"`
}
