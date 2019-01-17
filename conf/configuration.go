package conf

import "github.com/integration-system/isp-lib/structure"

type Configuration struct {
	ConfigServiceAddress structure.AddressConfiguration
	GrpcOuterAddress     structure.AddressConfiguration `valid:"required~Required" json:"grpcOuterAddress"`
	GrpcInnerAddress     structure.AddressConfiguration `valid:"required~Required" json:"grpcInnerAddress"`
	ModuleName           string
	InstanceUuid         string
}
