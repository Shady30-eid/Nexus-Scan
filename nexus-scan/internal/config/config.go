package config

import (
	"os"
	"strconv"
)

type Config struct {
	Version      string
	WSPort       int
	DatabasePath string
	LogLevel     string
	LogFile      string
	WorkerCount  int
	PollInterval int
}

func Load() *Config {
	cfg := &Config{
		Version:      "1.0.0",
		WSPort:       9999,
		DatabasePath: "./nexus-scan.db",
		LogLevel:     "info",
		LogFile:      "./nexus-scan.log",
		WorkerCount:  8,
		PollInterval: 3,
	}

	if v := os.Getenv("NEXUS_WS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.WSPort = port
		}
	}
	if v := os.Getenv("NEXUS_DB_PATH"); v != "" {
		cfg.DatabasePath = v
	}
	if v := os.Getenv("NEXUS_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("NEXUS_LOG_FILE"); v != "" {
		cfg.LogFile = v
	}
	if v := os.Getenv("NEXUS_WORKER_COUNT"); v != "" {
		if wc, err := strconv.Atoi(v); err == nil {
			cfg.WorkerCount = wc
		}
	}
	if v := os.Getenv("NEXUS_POLL_INTERVAL"); v != "" {
		if pi, err := strconv.Atoi(v); err == nil {
			cfg.PollInterval = pi
		}
	}

	return cfg
}
