// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package components

import (
	"context"
	"errors"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/mjpitz/myago/auth"
	basicauth "github.com/mjpitz/myago/auth/basic"
	oidcauth "github.com/mjpitz/myago/auth/oidc"
	"github.com/mjpitz/myago/headers"
)

type GRPCServerConfig struct {
	auth.Config
	OIDC  oidcauth.ClientConfig `json:"oidc"`
	Basic basicauth.Config      `json:"basic"`
}

func GRPCServer(ctx context.Context, cfg GRPCServerConfig) *grpc.Server {
	grpc_prometheus.EnableHandlingTimeHistogram()

	fns := []auth.HandlerFunc{
		func(ctx context.Context) (context.Context, error) {
			meta, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return ctx, nil
			}

			return headers.ToContext(ctx, headers.Header(meta)), nil
		},
	}

	switch cfg.Config.AuthType {
	case "basic":
		handler, err := basicauth.Handler(ctx, cfg.Basic)
		if err != nil {
			// don't do this...
			panic(err)
		}

		fns = append(fns, handler)
	case "oidc":
		fns = append(fns, oidcauth.OIDC(cfg.OIDC.Issuer))
	}

	fn := auth.Composite(fns...)
	authFunc := grpc_auth.AuthFunc(func(ctx context.Context) (context.Context, error) {
		ctx, err := fn(ctx)
		switch {
		case errors.Is(err, auth.ErrUnauthorized):
			return ctx, status.Errorf(codes.Unauthenticated, "unauthorized")
		case err != nil:
			return ctx, status.Errorf(codes.Internal, "internal error")
		}

		return ctx, nil
	})

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
