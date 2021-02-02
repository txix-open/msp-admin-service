package controller

import (
	"github.com/integration-system/isp-lib/v2/config"
	log "github.com/integration-system/isp-log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"msp-admin-service/conf"
)

func GetUIDesign(metadata metadata.MD) (conf.UIDesign, error) {
	token := metadata.Get(adminAuthHeaderName)

	if len(token) == 0 || token[0] == "" {
		log.Errorf(0, "Admin AUTH header: %s, not found, received: %v", adminAuthHeaderName, metadata)
		return conf.UIDesign{}, status.Error(codes.InvalidArgument, ServiceError)
	}

	return config.GetRemote().(*conf.RemoteConfig).UiDesign, nil
}
