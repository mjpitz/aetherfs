// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package components

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/mjpitz/aetherfs/internal/auth"
)

type GRPCServerConfig struct {
	AuthConfig auth.ClientConfig `json:""`
}

func GRPCServer(ctx context.Context, cfg GRPCServerConfig) *grpc.Server {
	grpc_prometheus.EnableHandlingTimeHistogram()

	var authFunc grpc_auth.AuthFunc

	switch cfg.AuthConfig.AuthType {
	case "oidc":
		authFunc = grpc_auth.AuthFunc(auth.Composite(
			auth.OIDCAuthenticator(cfg.AuthConfig.OIDC.Issuer),
			auth.RequireAuthentication(),
		))
	default:
		authFunc = grpc_auth.AuthFunc(auth.Composite())
	}

	server := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(ctxzap.Extract(ctx)),
			grpc_auth.UnaryServerInterceptor(authFunc),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(ctxzap.Extract(ctx)),
			grpc_auth.StreamServerInterceptor(authFunc),
		),
	)

	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	reflection.Register(server)

	return server
}
