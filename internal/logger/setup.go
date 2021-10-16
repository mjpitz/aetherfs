// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package logger

import (
	"context"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config encapsulates the configurable elements of the logger.
type Config struct {
	Level  string `json:"level"  usage:"adjust the verbosity of the logs" default:"info"`
	Format string `json:"format" usage:"configure the format of the logs" default:"json"`
}

// Setup creates a logger given the provided configuration.
func Setup(ctx context.Context, cfg Config) context.Context {
	level := zapcore.InfoLevel
	if cfg.Level != "" {
		err := (&level).Set(cfg.Level)
		if err != nil {
			panic(err)
		}
   	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.Level.SetLevel(level)
	zapConfig.Encoding = cfg.Format
	zapConfig.Sampling = nil // don't sample

	logger, err := zapConfig.Build()
	if err != nil {
		panic(err)
	}

	// silence grpc logs for now
	grpcLogger := logger
	if level > zapcore.DebugLevel {
		grpcLogger = zap.NewNop()
	}
	grpc_zap.ReplaceGrpcLoggerV2(grpcLogger)

	return ctxzap.ToContext(ctx, logger)
}
