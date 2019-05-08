package conf

import (
	"github.com/integration-system/isp-lib/structure"
)

type RemoteConfig struct {
	Database  structure.DBConfiguration     `valid:"required~Required" json:"database" schema:"Database"`
	Metrics   structure.MetricConfiguration `schema:"Metrics"`
	SecretKey string                        `valid:"required~Required" schema:"JWT secret"`
	ExpireSec int                           `schema:"Token expire time,in seconds"`
}
