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
	"os/signal"
	"gitlab8.alx/msp2.0/msp-lib/database"
)

var (
	configData        *conf.Configuration
	executableFileDir string
	configChan        = make(chan bool)
	version           = "0.1.0"
	date              = "undefined"
)

func init() {
	config.InitConfig(&conf.Configuration{})
	validLocalConfig()
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
		logger.Warnf("Remote config isn't received, module: %s, instance_uuid: %s",
			appConfig.ModuleName, appConfig.InstanceUuid)
		time.Sleep(time.Second * 5)
	}
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
	listenConfigChange()
	awaitTerminate()
}

// Start a GRPC server.
func createGrpcServer() {
	remoteConfig := config.GetRemote().(*conf.RemoteConfig)
	addr := structure.AddressConfiguration{IP: remoteConfig.GrpcAddress.IP, Port: remoteConfig.GrpcAddress.Port}
	backend.StartBackendGrpcServer(addr, backend.GetDefaultService(remoteConfig.GrpcPrefix, helper.GetHandlers()))
}

func validRemoteConfig(remoteConfig *conf.RemoteConfig) {
	_, err := govalidator.ValidateStruct(remoteConfig)
	if err != nil {
		validationErrors := govalidator.ErrorsByField(err)
		logger.Fatal("Remote config int't valid", validationErrors)
		panic(err)
	}
}

func validLocalConfig() {
	localConfig := config.Get().(*conf.Configuration)
	_, err := govalidator.ValidateStruct(localConfig)
	if err != nil {
		validationErrors := govalidator.ErrorsByField(err)
		logger.Fatal("Local config int't valid", validationErrors)
		panic(err)
	}
}

func subscribeSocket(client *gosocketio.Client, eventName string) {
	client.On(eventName, func(h *gosocketio.Channel, args string) error {
		logger.Infof("--- Got event: %s message: %s", eventName, args)
		remoteConfig := &conf.RemoteConfig{}
		config.InitRemoteConfig(remoteConfig, args)
		validRemoteConfig(remoteConfig)
		configChan <- true
		return nil
	})
}

func listenConfigChange() {
	go func() {
		for {
			<-configChan
			
			localConfig := config.Get().(*conf.Configuration)
			remoteConfig := config.GetRemote().(*conf.RemoteConfig)
			
			backend.StopGrpcServer()
			createGrpcServer()
			database.InitDb(remoteConfig.Database)
			
			addrOuter := structure.AddressConfiguration{
				IP:   localConfig.GrpcOuterAddress.IP,
				Port: localConfig.GrpcOuterAddress.Port,
			}
			
			methods := backend.CreateBackendConfig(
				structure.ModuleInfo{ModuleName: localConfig.ModuleName, Version: version},
				addrOuter,
				remoteConfig.GrpcPrefix, helper.GetHandlers())
			
			bytes, err := json.Marshal(methods)
			if err != nil {
				logger.Warn("Error when serializing Backend Routes", err)
			} else {
				logger.Infof("EXPORTED MODULE METHODS: %s", methods)
				socketClient := socket.GetClient()
				socketClient.Emit(utils.SendRoutesWhenConnected, string(bytes))
			}
			
		}
	}()
}

func awaitTerminate() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	logger.Info("Shutting down")
	os.Exit(0)
}
