package service

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	etp "github.com/integration-system/isp-etp-go/v2"
	"github.com/integration-system/isp-lib/v2/structure"
	log "github.com/integration-system/isp-log"
	"msp-admin-service/invoker"
)

const (
	onConfigInEvent        = "config_in"
	onConfigOutEvent       = "config_out"
	onConfigUpdateEvent    = "config_upd"
	onModuleUpdateEvent    = "module_upd"
	onUniversalUpdateEvent = "universal_upd"

	broadcastConfigUpdateEvent = "force_config_update"
	broadcastModuleUpdateEvent = "force_module_update"
)

type (
	sessionManager struct {
		lock       sync.Mutex
		session    map[string]sessionInfo
		wsServer   etp.Server
		httpServer *http.Server
	}
	sessionInfo struct {
		ConfigID string `json:"id"`
		ModuleID string `json:"module"`
		Version  int    `json:"version"`
	}
	sessionSend struct {
		Module string        `json:"module"`
		Data   []sessionInfo `json:"data"`
	}
)

var SessionManager = sessionManager{
	session: make(map[string]sessionInfo),
}

func (s *sessionManager) GetConfSessionByModule(id string) []sessionInfo {
	var data []sessionInfo
	s.lock.Lock()
	for _, v := range s.session {
		if v.ModuleID == id {
			data = append(data, v)
		}
	}
	s.lock.Unlock()

	return data
}

func (s *sessionManager) AddConfSession(data []byte, id string) {
	var sessionData sessionInfo
	if err := json.Unmarshal(data, &sessionData); err != nil {
		log.Error(44, "get date from request: This session can't be registered")
		return
	}
	s.lock.Lock()
	s.session[id] = sessionData
	s.lock.Unlock()

	send := sessionSend{
		Module: sessionData.ModuleID,
		Data:   s.GetConfSessionByModule(sessionData.ModuleID),
	}

	if err := s.wsServer.BroadcastToAll(broadcastConfigUpdateEvent, jsoner(send)); err != nil {
		log.Error(44, "send new session info to clients")
	}
}

func (s *sessionManager) DropConfSession(id string) {
	var modID string
	s.lock.Lock()
	if value, ok := s.session[id]; !ok {
		s.lock.Unlock()
		return
	} else {
		modID = value.ModuleID
	}

	delete(s.session, id)
	s.lock.Unlock()

	send := sessionSend{
		Module: modID,
		Data:   s.GetConfSessionByModule(modID),
	}

	if err := s.wsServer.BroadcastToAll(broadcastConfigUpdateEvent, jsoner(send)); err != nil {
		log.Error(44, "send new session info to clients")
	}
}

func jsoner(data interface{}) []byte {
	byt, _ := json.Marshal(data)
	return byt
}

func (s *sessionManager) updateModulesList() bool {
	result := invoker.GetModulesInfo()
	if result != nil {
		if err := s.wsServer.BroadcastToAll(broadcastModuleUpdateEvent, jsoner(result)); err != nil {
			log.Error(44, "sending new module info to clients")
		}
	}

	return true
}

func (s *sessionManager) RoutesUpdateSessionCallback(_ structure.RoutingConfig) bool {
	return s.updateModulesList()
}

func (s *sessionManager) InitWebSocket(ln net.Listener) {
	etpServerConfig := etp.ServerConfig{
		InsecureSkipVerify: true,
	}

	s.wsServer = etp.NewServer(context.TODO(), etpServerConfig).
		OnDisconnect(func(conn etp.Conn, err error) {
			s.DropConfSession(conn.ID())
		}).
		OnError(func(conn etp.Conn, err error) {
			s.DropConfSession(conn.ID())
			log.Errorf(44, "On WebSocket client connection error occurred: %v", err)
		}).
		//on open session
		On(onConfigInEvent, func(conn etp.Conn, data []byte) {
			s.AddConfSession(data, conn.ID())
		}).
		//on close session
		On(onConfigOutEvent, func(conn etp.Conn, data []byte) {
			s.DropConfSession(conn.ID())
		}).
		//on update data session
		On(onUniversalUpdateEvent, func(conn etp.Conn, data []byte) {
			var (
				responseData json.RawMessage
				requestData  invoker.ConfigRequest
			)

			if err := json.Unmarshal(data, &requestData); err != nil {
				log.Errorf(44, "parse request data: %v", err)
				return
			}

			if responseData = invoker.GetConfigsById(requestData); responseData == nil {
				return
			}

			send := struct {
				ModuleId string `json:"moduleId"`
				//Data     json.RawMessage `json:"data"`
			}{
				ModuleId: requestData.ModuleId,
				//Data:     responseData,
			}

			if err := s.wsServer.BroadcastToAll("force_update", jsoner(send)); err != nil {
				log.Errorf(44, "send new config info for module with id %v : %v", requestData.ModuleId, err)
			}
		}).
		//on session info request (on mount config page)
		On(onConfigUpdateEvent, func(conn etp.Conn, data []byte) {
			var request struct {
				Module string `json:"module"`
			}

			if err := json.Unmarshal(data, &request); err != nil {
				return
			}

			modID := request.Module

			send := sessionSend{
				Module: modID,
				Data:   s.GetConfSessionByModule(modID),
			}
			if err := conn.Emit(context.TODO(), broadcastConfigUpdateEvent, jsoner(send)); err != nil {
				log.Errorf(44, "send new session info to clients : %v", err)
			}
		}).
		//on modules info request (on delete module in module list (client))
		On(onModuleUpdateEvent, func(conn etp.Conn, data []byte) {
			s.updateModulesList()
		})

	mux := http.NewServeMux()
	mux.HandleFunc("/admin/notify", s.wsServer.ServeHttp)
	s.httpServer = &http.Server{Handler: mux}
	go func() {
		if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf(44, "Unable to start http server. err: %v", err)
		}
	}()
}

func (s *sessionManager) ShutdownSocket() {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Errorf(44, "Unable to shutdown http server. err: %v", err)
	}

	s.wsServer.Close()
}
