package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nexus-scan/nexus-scan/internal/config"
	"github.com/nexus-scan/nexus-scan/pkg/device"
	"github.com/nexus-scan/nexus-scan/pkg/ipc"
	"github.com/nexus-scan/nexus-scan/pkg/logger"
	"github.com/nexus-scan/nexus-scan/pkg/remediation"
	"github.com/nexus-scan/nexus-scan/pkg/scanner"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "" ||
			origin == "tauri://localhost" ||
			origin == "http://localhost" ||
			origin == "https://tauri.localhost"
	},
}

type WebSocketServer struct {
	cfg            *config.Config
	broadcaster    *ipc.Broadcaster
	scannerSvc     *scanner.Scanner
	remediationSvc *remediation.Service
	deviceMgr      *device.Manager
	log            *logger.Logger
	httpServer     *http.Server
	mu             sync.Mutex
}

func NewWebSocketServer(
	cfg *config.Config,
	broadcaster *ipc.Broadcaster,
	scannerSvc *scanner.Scanner,
	remediationSvc *remediation.Service,
	deviceMgr *device.Manager,
	log *logger.Logger,
) *WebSocketServer {
	return &WebSocketServer{
		cfg:            cfg,
		broadcaster:    broadcaster,
		scannerSvc:     scannerSvc,
		remediationSvc: remediationSvc,
		deviceMgr:      deviceMgr,
		log:            log,
	}
}

func (s *WebSocketServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)

	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", s.cfg.WSPort),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      0,
		ReadTimeout:       0,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("http server failed: %w", err)
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

func (s *WebSocketServer) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *WebSocketServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "version": s.cfg.Version})
}

func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Error("websocket upgrade failed", "error", err, "remote", r.RemoteAddr)
		return
	}

	client := ipc.NewClient(conn, s.log)
	s.broadcaster.Register(client)
	defer func() {
		s.broadcaster.Unregister(client)
		conn.Close()
	}()

	s.log.Info("websocket client connected", "remote", r.RemoteAddr)

	go client.WritePump()

	s.sendWelcome(client)

	s.readPump(conn, client)
}

func (s *WebSocketServer) sendWelcome(client *ipc.Client) {
	welcome := ipc.Event{
		Type:      "connected",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Payload: map[string]interface{}{
			"version": s.cfg.Version,
			"message": "Nexus-Scan backend connected",
		},
	}
	client.Send(welcome)
}

func (s *WebSocketServer) readPump(conn *websocket.Conn, client *ipc.Client) {
	conn.SetReadLimit(65536)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.log.Error("websocket read error", "error", err)
			}
			return
		}
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		var cmd ipc.Command
		if err := json.Unmarshal(message, &cmd); err != nil {
			s.log.Warn("invalid command received", "error", err)
			client.Send(ipc.ErrorEvent("invalid_json", "malformed command payload"))
			continue
		}

		s.handleCommand(cmd, client)
	}
}

func (s *WebSocketServer) handleCommand(cmd ipc.Command, client *ipc.Client) {
	s.log.Info("command received", "type", cmd.Type, "payload", cmd.Payload)

	switch cmd.Type {
	case "start_scan":
		deviceID, ok := cmd.Payload["device_id"].(string)
		if !ok || deviceID == "" {
			client.Send(ipc.ErrorEvent("invalid_payload", "device_id required"))
			return
		}
		go s.scannerSvc.StartScan(deviceID)

	case "stop_scan":
		deviceID, ok := cmd.Payload["device_id"].(string)
		if !ok || deviceID == "" {
			client.Send(ipc.ErrorEvent("invalid_payload", "device_id required"))
			return
		}
		s.scannerSvc.StopScan(deviceID)

	case "remediate":
		deviceID, ok := cmd.Payload["device_id"].(string)
		packageName, ok2 := cmd.Payload["package_name"].(string)
		if !ok || !ok2 || deviceID == "" || packageName == "" {
			client.Send(ipc.ErrorEvent("invalid_payload", "device_id and package_name required"))
			return
		}
		go s.remediationSvc.Uninstall(deviceID, packageName)

	case "list_devices":
		devices := s.deviceMgr.ListDevices()
		client.Send(ipc.Event{
			Type:      "device_list",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Payload: map[string]interface{}{
				"devices": devices,
			},
		})

	case "get_scan_history":
		deviceID, _ := cmd.Payload["device_id"].(string)
		history, err := s.scannerSvc.GetScanHistory(deviceID)
		if err != nil {
			client.Send(ipc.ErrorEvent("db_error", err.Error()))
			return
		}
		client.Send(ipc.Event{
			Type:      "scan_history",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Payload: map[string]interface{}{
				"history": history,
			},
		})

	case "ping":
		client.Send(ipc.Event{
			Type:      "pong",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Payload:   map[string]interface{}{},
		})

	default:
		s.log.Warn("unknown command type", "type", cmd.Type)
		client.Send(ipc.ErrorEvent("unknown_command", fmt.Sprintf("unknown command: %s", cmd.Type)))
	}
}
