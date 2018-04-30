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
	"net"
	"gitlab8.alx/msp2.0/msp-lib/database"
)

var (
	configData        *conf.Configuration
	executableFileDir string
	methodBytes       []byte
)
var socketConnected = false

func init() {
	config.InitConfig(&conf.Configuration{})
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
			socket.GetClient().On("connection", func(so *gosocketio.Channel) {
				socketConnected = true
				if methodBytes != nil {
					so.Emit(utils.SendRoutesWhenConnected, string(methodBytes))
				}
			})
		},
	)
	time.Sleep(time.Second * 3)
	for config.GetRemote() == nil {
		logger.Warn("Remote config isn't received")
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
	//createGrpcServer()
	awaitTerminate()
}

// Start a GRPC server.
func createGrpcServer() {
	// Run our server in a goroutine so that it doesn't block.
	handlers := helper.GetHandlers()
	remoteConfig := config.GetRemote().(*conf.RemoteConfig)
	appConfig := config.Get().(*conf.Configuration)
	addr := structure.AddressConfiguration{IP: remoteConfig.GrpcAddress.IP, Port: remoteConfig.GrpcAddress.Port}
	addrOuter := structure.AddressConfiguration{
		IP: appConfig.GrpcOuterAddress.IP,
		Port: appConfig.GrpcOuterAddress.Port,
	}
	backend.StartBackendGrpcServer(addr, backend.GetDefaultService(remoteConfig.GrpcPrefix, handlers))
	methods := backend.GetBackendConfig(addrOuter, remoteConfig.GrpcPrefix, handlers)
	bytes, err := json.Marshal(methods)
	if err != nil {
		logger.Warn("Error when serializing Backend Routes", err)
	} else {
		methodBytes = bytes
		logger.Infof("EXPORTED MODULE METHODS: %s", methods)
		if socketConnected {
			socket.GetClient().Emit(utils.SendRoutesWhenConnected, string(bytes))
		}
	}
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
		database.InitDb(remoteConfig.Database)
		backend.StopGrpcServer()
		for !checkPortIsFree(remoteConfig.GrpcAddress.Port) {
			time.Sleep(time.Second * 3)
			logger.Info("Wait for free port for new grpc connection")
		}
		createGrpcServer()
		return nil
	})
}

func checkPortIsFree(port string) bool {
	ln, err := net.Listen("tcp", ":" + port)
	if err != nil {
		return false
	} else {
		defer ln.Close()
		return true
	}
}

func awaitTerminate() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	logger.Info("Shutting down")
	os.Exit(0)
}
