package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
	base *zap.Logger
}

func New(level string, logFile string) (*Logger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "module",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)
	jsonEncoder := zapcore.NewJSONEncoder(encoderCfg)

	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stderr), zapLevel)

	var cores []zapcore.Core
	cores = append(cores, consoleCore)

	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			fileCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(f), zapLevel)
			cores = append(cores, fileCore)
		}
	}

	combined := zapcore.NewTee(cores...)
	base := zap.New(combined, zap.AddCaller(), zap.AddCallerSkip(0))

	return &Logger{
		SugaredLogger: base.Sugar(),
		base:          base,
	}, nil
}

func (l *Logger) WithDevice(deviceID string) *Logger {
	child := l.base.With(zap.String("device_id", deviceID))
	return &Logger{
		SugaredLogger: child.Sugar(),
		base:          child,
	}
}

func (l *Logger) WithModule(module string) *Logger {
	child := l.base.With(zap.String("module", module))
	return &Logger{
		SugaredLogger: child.Sugar(),
		base:          child,
	}
}

func (l *Logger) ThreatEvent(deviceID, packageName, threatName, severity string) {
	l.base.Info("threat detected",
		zap.String("level", "THREAT"),
		zap.String("device_id", deviceID),
		zap.String("package_name", packageName),
		zap.String("threat_name", threatName),
		zap.String("threat_level", severity),
		zap.String("action", "flagged"),
	)
}

func (l *Logger) RemediationEvent(deviceID, packageName, action string, success bool) {
	l.base.Info("remediation action",
		zap.String("level", "INFO"),
		zap.String("device_id", deviceID),
		zap.String("package_name", packageName),
		zap.String("action", action),
		zap.Bool("success", success),
	)
}

func (l *Logger) Sync() error {
	return l.base.Sync()
}
