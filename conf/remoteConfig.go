package conf

import (
	"github.com/integration-system/isp-lib/structure"
)

type RemoteConfig struct {
	Database structure.DBConfiguration     `valid:"required~Required" json:"database" schema:"Database"`
	Metrics  structure.MetricConfiguration `schema:"Metrics"`
}
