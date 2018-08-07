package conf

import "gitlab.alx/msp2.0/msp-lib/database"

type RemoteConfig struct {
	Database database.DBConfiguration `valid:"required~Required" json:"database"`
}
