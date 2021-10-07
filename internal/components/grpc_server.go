// AetherFS - A virtual file system for small to medium sized datasets (MB or GB, not TB or PB).
// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
