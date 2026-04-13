package logger

import (
	"strings"

	"go.uber.org/zap"
)

// NewLogger создаёт и настраивает логгер в зависимости от уровня логирования.
func NewLogger(level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()

	switch strings.ToLower(level) {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	return cfg.Build()
}
