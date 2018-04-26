package conf

type Configuration struct {
	ConfigServiceAddress AddressConfiguration
	GrpcOuterAddress     AddressConfiguration `valid:"required~Required" json:"grpcOuterAddress"`
	ModuleName           string
	InstanceUuid         string
}

type AddressConfiguration struct {
	Port string `valid:"required~Required"`
	IP   string `valid:"required~Required"`
}

func (addressConfiguration *AddressConfiguration) GetAddress() string {
	return addressConfiguration.IP + ":" + addressConfiguration.Port
}
