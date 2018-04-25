package conf

import "gitlab8.alx/msp2.0/msp-lib/database"

type RemoteConfig struct {
	GrpcInnerIp string                   `valid:"required~Required"`
	GrpcPrefix  string
	GrpcAddress AddressConfiguration     `valid:"required~Required"`
	Database    database.DBConfiguration `valid:"required~Required"`
}
