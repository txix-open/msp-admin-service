package main

import (
	rn "runtime"
	"gitlab8.alx/msp2.0/msp-lib/backend"
	"gitlab8.alx/msp2.0/msp-lib/config"
	"gitlab8.alx/msp2.0/msp-lib/socket"
	"gitlab8.alx/msp2.0/msp-lib/logger"
	"gitlab8.alx/msp2.0/msp-lib/structure"
	"gitlab8.alx/msp2.0/msp-lib/utils"
	"admin-service/conf"
	"os"
	"path"
	"admin-service/helper"
	"time"
	"github.com/asaskevich/govalidator"
	"github.com/graarh/golang-socketio"
	"encoding/json"
	"gitlab8.alx/msp2.0/msp-lib/database"
)

var (
	configData        *conf.Configuration
	executableFileDir string
)

func init() {
	appConfig := config.Get().(*conf.Configuration)
	socket.InitClient(
		socket.SocketConfiguration{
			Host:      appConfig.ConfigServiceAddress.IP,
			Port:      appConfig.ConfigServiceAddress.Port,
			Secure:    false,
			UrlParams: map[string]string{"module_name": appConfig.ModuleName, "instance_uuid": appConfig.InstanceUuid},
		},
		func(client *gosocketio.Client) {
			subscribeSocket(client, utils.SendConfigWhenConnected)
			subscribeSocket(client, utils.SendConfigChanged)
			subscribeSocket(client, utils.SendConfigOnRequest)
		},
	)
	time.Sleep(time.Second * 3)
	for config.GetRemote() == nil {
		logger.Warn("Remote config isn't received")
		time.Sleep(time.Second * 5)
	}
	remoteConfig := config.GetRemote().(*conf.RemoteConfig)
	validRemoteConfig(remoteConfig)
	config.InitConfig(&conf.Configuration{})
	database.InitDb(remoteConfig.Database)
}

func main() {
	ex, err := os.Executable()
	executableFileDir = path.Dir(ex)
	if err != nil {
		logger.Fatal(err)
		panic(err)
	}
	if utils.DEV {
		_, filename, _, _ := rn.Caller(0)
		executableFileDir = path.Dir(filename)
	}
	createGrpcServer()
}

// Start a GRPC server.
func createGrpcServer() {
	// Run our server in a goroutine so that it doesn't block.
	handlers := helper.GetHandlers()
	remoteConfig := config.GetRemote().(*conf.RemoteConfig)
	addr := structure.AddressConfiguration{IP: remoteConfig.GrpcAddress.IP, Port: remoteConfig.GrpcAddress.Port}
	addrInner := structure.AddressConfiguration{IP: remoteConfig.GrpcInnerIp, Port: remoteConfig.GrpcAddress.Port}
	backend.StartBackendGrpcServer(addr, backend.GetDefaultService(remoteConfig.GrpcPrefix+"/", handlers))
	methods := backend.GetBackendConfig(addrInner, remoteConfig.GrpcPrefix+"/", handlers)
	bytes, err := json.Marshal(methods)
	if err != nil {
		logger.Warn("Error when serializing Backend Routes", err)
	}
	socket.GetClient().Emit(utils.SendRoutesWhenConnected, string(bytes))
	logger.Infof("EXPORTED MODULE METHODS: %s", methods)
}

func validRemoteConfig(remoteConfig *conf.RemoteConfig) {
	_, err := govalidator.ValidateStruct(remoteConfig)
	if err != nil {
		validationErrors := govalidator.ErrorsByField(err)
		logger.Fatal("Remote config int't valid", validationErrors)
		panic(err)
	}
}

func subscribeSocket(client *gosocketio.Client, eventName string) {
	client.On(eventName, func(h *gosocketio.Channel, args string) error {
		logger.Infof("--- Got event: %s message: %s", eventName, args)
		remoteConfig := &conf.RemoteConfig{}
		config.InitRemoteConfig(remoteConfig, args)
		validRemoteConfig(remoteConfig)
		return nil
	})
}