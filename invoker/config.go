package invoker

import (
	"encoding/json"
	"time"

	"github.com/integration-system/isp-lib/v2/config/schema"
	log "github.com/integration-system/isp-log"

	"github.com/integration-system/isp-lib/v2/backend"
	"google.golang.org/grpc"
)

type (
	ModuleInfo struct {
		Id                 string         `json:"id"`
		Name               string         `json:"name"`
		Active             bool           `json:"active"`
		CreatedAt          time.Time      `json:"createdAt"`
		LastConnectedAt    time.Time      `json:"lastConnectedAt"`
		LastDisconnectedAt time.Time      `json:"lastDisconnectedAt"`
		ConfigSchema       *schema.Schema `json:"configSchema"`
		Status             interface{}    `json:"status"`
		RequiredModules    interface{}    `json:"requiredModules"`
	}

	ConfigRequest struct {
		ModuleId string `json:"moduleId"`
	}
)

var (
	ConfigClient = backend.NewRxGrpcClient(
		backend.WithDialOptions(grpc.WithInsecure()),
	)
)

func GetConfigsById(request ConfigRequest) json.RawMessage {
	var response json.RawMessage
	if err := ConfigClient.Invoke("config/config/get_configs_by_module_id", 0, request, &response); err != nil {
		log.Errorf(77, "can't get response from grpc request : %v", err)
		return nil
	}

	return response
}

func GetModulesInfo() []ModuleInfo {
	var response []ModuleInfo
	if err := ConfigClient.Invoke("config/module/get_modules_info", 0, nil, &response); err != nil {
		log.Errorf(77, "can't get response from grpc request : %v", err)
		return nil
	}

	return response
}
