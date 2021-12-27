// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package components

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"

	"github.com/mjpitz/aetherfs/internal/storage/local"
	"github.com/mjpitz/myago/auth"
	basicauth "github.com/mjpitz/myago/auth/basic"
	oidcauth "github.com/mjpitz/myago/auth/oidc"
	"github.com/mjpitz/myago/lifecycle"
	"github.com/mjpitz/myago/livetls"
)

const defaultServiceConfig = `{
  "loadBalancingPolicy": "round_robin",
  "healthCheckConfig": {
    "serviceName": ""
  }
}`

type GRPCClientConfig struct {
	Target string         `json:"target" usage:"address the grpc client should dial"`
	TLS    livetls.Config `json:"tls"`

	auth.Config
	OIDC  oidcauth.Config        `json:"oidc"`
	Basic basicauth.ClientConfig `json:"basic"`
}

func GRPCClient(ctx context.Context, cfg GRPCClientConfig) (*grpc.ClientConn, error) {
	tokens := local.Extract(ctx).Tokens()

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

	tlsConfig, err := livetls.New(ctx, cfg.TLS)
	if err != nil {
		return nil, err
	}

	if tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	var tokenSource oauth2.TokenSource

	switch cfg.AuthType {
	case "basic":
		token, err := cfg.Basic.Token()
		if err != nil {
			return nil, err
		}

		tokenSource = oauth2.StaticTokenSource(token)
	case "oidc":
		token := &oauth2.Token{}
		err := tokens.Get(ctx, cfg.Target, token)
		if err != nil {
			return nil, err
		}

		tokenSource = oauth2.StaticTokenSource(token)
	}

	if tokenSource != nil {
		opts = append(opts, grpc.WithPerRPCCredentials(
			oauth.TokenSource{
				TokenSource: tokenSource,
			}),
		)
	}

	cc, err := grpc.Dial(cfg.Target, opts...)
	if err != nil {
		return nil, err
	}

	lifecycle.Defer(func(ctx context.Context) {
		_ = cc.Close()
	})

	return cc, nil
}
