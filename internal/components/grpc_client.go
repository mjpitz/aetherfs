// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package components

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/mjpitz/aetherfs/internal/lifecycle"
)

const defaultServiceConfig = `{
  "loadBalancingPolicy": "round_robin",
  "healthCheckConfig": {
    "serviceName": ""
  }
}`

type GRPCClientConfig struct {
	Target    string    `json:"target" usage:"address the grpc client should dial"`
	TLSConfig TLSConfig `json:"tls"`
}

func GRPCClient(ctx context.Context, cfg GRPCClientConfig) (*grpc.ClientConn, error) {
	grpc_prometheus.EnableClientHandlingTimeHistogram()

	backoff := grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond))

	unaryInterceptors := []grpc.UnaryClientInterceptor{
		grpc_retry.UnaryClientInterceptor(backoff),
		grpc_prometheus.UnaryClientInterceptor,
		grpc_zap.UnaryClientInterceptor(ctxzap.Extract(ctx)),
	}

	streamInterceptors := []grpc.StreamClientInterceptor{
		grpc_prometheus.StreamClientInterceptor,
		grpc_zap.StreamClientInterceptor(ctxzap.Extract(ctx)),
	}

	opts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(defaultServiceConfig),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(unaryInterceptors...)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(streamInterceptors...)),
	}

	tlsConfig, err := LoadCertificates(cfg.TLSConfig)
	if err != nil {
		return nil, err
	}

	if tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	cc, err := grpc.Dial(cfg.Target, opts...)
	if err != nil {
		return nil, err
	}

	lifecycle.Defer(func(ctx context.Context) {
		err = cc.Close()
		if err != nil {
			ctxzap.Extract(ctx).Error("failed to close grpc client connection", zap.Error(err))
		}
	})

	return cc, nil
}
