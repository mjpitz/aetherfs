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
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/mjpitz/aetherfs/internal/lifecycle"
)

type HTTPServerConfig struct {
	Port      int       `json:"port" usage:"which port the HTTP server should be bound to"`
	TLSConfig TLSConfig `json:"tls"`
}

func ListenAndServeHTTP(ctx context.Context, cfg HTTPServerConfig, handler http.Handler) error {
	tlsConfig, err := LoadCertificates(cfg.TLSConfig)
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
		Addr:    fmt.Sprintf("0.0.0.0:%d", cfg.Port),
		Handler: handler,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
		TLSConfig: tlsConfig,
	}

	lifecycle.Defer(func(ctx context.Context) {
		err = svr.Shutdown(ctx)
	})

	go func() {
		_ = svr.ListenAndServe()
	}()

	return nil
}
