package conf

import "gitlab8.alx/msp2.0/msp-lib/database"

type RemoteConfig struct {
	GrpcPrefix       string                   `json:"grpcPrefix"`
	GrpcAddress      AddressConfiguration     `valid:"required~Required" json:"grpcAddress"`
	Database         database.DBConfiguration `valid:"required~Required" json:"database"`
}
