package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nexus-scan/nexus-scan/internal/config"
	"github.com/nexus-scan/nexus-scan/internal/server"
	"github.com/nexus-scan/nexus-scan/pkg/database"
	"github.com/nexus-scan/nexus-scan/pkg/device"
	"github.com/nexus-scan/nexus-scan/pkg/ipc"
	"github.com/nexus-scan/nexus-scan/pkg/logger"
	"github.com/nexus-scan/nexus-scan/pkg/remediation"
	"github.com/nexus-scan/nexus-scan/pkg/scanner"
)

func main() {
	cfg := config.Load()

	log, err := logger.New(cfg.LogLevel, cfg.LogFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("nexus-scan starting", "version", cfg.Version, "port", cfg.WSPort)

	db, err := database.New(cfg.DatabasePath, log)
	if err != nil {
		log.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Error("failed to run database migrations", "error", err)
		os.Exit(1)
	}

	if err := db.Seed(); err != nil {
		log.Warn("failed to seed database (non-fatal)", "error", err)
	}

	broadcaster := ipc.NewBroadcaster(log)
	scannerSvc := scanner.New(db, broadcaster, log, cfg.WorkerCount)
	remediationSvc := remediation.New(db, broadcaster, log)
	deviceMgr := device.NewManager(broadcaster, scannerSvc, log)

	wsServer := server.NewWebSocketServer(cfg, broadcaster, scannerSvc, remediationSvc, deviceMgr, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go deviceMgr.StartPolling(ctx)

	if err := wsServer.Start(ctx); err != nil {
		log.Error("failed to start websocket server", "error", err)
		os.Exit(1)
	}

	log.Info("nexus-scan ready", "ws_address", fmt.Sprintf("ws://127.0.0.1:%d/ws", cfg.WSPort))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down nexus-scan...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := wsServer.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}

	log.Info("nexus-scan stopped")
}
