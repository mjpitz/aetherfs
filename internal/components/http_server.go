// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package components

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/mjpitz/myago/lifecycle"
	"github.com/mjpitz/myago/livetls"
)

type HTTPServerConfig struct {
	Port int            `json:"port" usage:"which port the HTTP server should be bound to" default:"8080"`
	TLS  livetls.Config `json:"tls"`
}

func ListenAndServeHTTP(ctx context.Context, cfg HTTPServerConfig, handler http.Handler) error {
	tlsConfig, err := livetls.New(ctx, cfg.TLS)
	if err != nil {
		return err
	}

	if tlsConfig != nil && len(tlsConfig.Certificates) > 0 {
		// enforce mutual TLS
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	handler = cors.Default().Handler(handler)
	handler = h2c.NewHandler(handler, &http2.Server{})

	svr := &http.Server{
		Handler: handler,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
		TLSConfig: tlsConfig,
	}

	lifecycle.Defer(func(ctx context.Context) {
		err = svr.Shutdown(ctx)
	})

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.Port))
	if err != nil {
		return err
	}

	if l != nil && tlsConfig != nil {
		l = tls.NewListener(l, tlsConfig)
	}

	go func() {
		_ = svr.Serve(l)
	}()

	return nil
}
