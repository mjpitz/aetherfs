// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

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
	Level  *zapcore.Level `json:"level,omitempty"  usage:"adjust the verbosity of the logs"`
	Format string         `json:"format,omitempty" usage:"configure the format of the logs"`
}

// Setup creates a logger given the provided configuration.
func Setup(ctx context.Context, cfg Config) context.Context {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level.SetLevel(*(cfg.Level))
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
