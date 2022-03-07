package cli

import (
	"fmt"

	"go.uber.org/zap"
)

func mustInitializeLogger(level zap.AtomicLevel) *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = level
	cfg.Development = false
	logger, err := cfg.Build()
	if err != nil {
		panic(fmt.Errorf("failed to initialize logger: %v", err))
	}
	return logger
}
