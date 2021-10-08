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
	Level  string `json:"level"  usage:"adjust the verbosity of the logs"`
	Format string `json:"format" usage:"configure the format of the logs"`
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

	grpc_zap.ReplaceGrpcLogger(logger)
	grpc_zap.ReplaceGrpcLoggerV2(logger)

	return ctxzap.ToContext(ctx, logger)
}
