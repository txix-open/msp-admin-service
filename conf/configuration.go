package conf

import "gitlab.alx/msp2.0/msp-lib/structure"

type Configuration struct {
	ConfigServiceAddress structure.AddressConfiguration
	GrpcOuterAddress     structure.AddressConfiguration `valid:"required~Required" json:"grpcOuterAddress"`
	GrpcInnerAddress     structure.AddressConfiguration `valid:"required~Required" json:"grpcInnerAddress"`
	ModuleName           string
	InstanceUuid         string
}
