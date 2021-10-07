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

package run

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/urfave/cli/v2"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	"github.com/mjpitz/aetherfs/internal/components"
	"github.com/mjpitz/aetherfs/internal/flagset"
	"github.com/mjpitz/aetherfs/internal/fs"
	"github.com/mjpitz/aetherfs/internal/storage"
	"github.com/mjpitz/aetherfs/internal/storage/s3"
)

// HubConfig encapsulates the requirements for configuring and starting up the Hub process.
type HubConfig struct {
	HTTPServerConfig components.HTTPServerConfig `json:""`
	GRPCServerConfig components.GRPCServerConfig `json:""`
	StorageConfig    storage.Config              `json:"storage"`
}

// Hub returns a command that will run the server process.
func Hub() *cli.Command {
	cfg := &HubConfig{
		HTTPServerConfig: components.HTTPServerConfig{
			Port: 8080,
		},
		StorageConfig: storage.Config{
			Driver: "s3",
			S3: s3.Config{
				Endpoint: "s3.amazonaws.com",
				TLS:      components.TLSConfig{},
			},
		},
	}

	return &cli.Command{
		Name:        "hub",
		Usage:       "Runs the AetherFS Hub process",
		UsageText:   "aetherfs run hub [options]",
		Description: "The aetherfs-hub process is responsible for collecting and hosting datasets.",
		Flags:       flagset.Extract(cfg),
		Action: func(ctx *cli.Context) error {
			serverConn, err := components.GRPCClient(ctx.Context, components.GRPCClientConfig{
				Target:    fmt.Sprintf("localhost:%d", cfg.HTTPServerConfig.Port),
				TLSConfig: cfg.HTTPServerConfig.TLSConfig,
			})
			if err != nil {
				return err
			}

			blockAPI := blockv1.NewBlockAPIClient(serverConn)
			datasetAPI := datasetv1.NewDatasetAPIClient(serverConn)

			stores, err := storage.ObtainStores(cfg.StorageConfig)
			if err != nil {
				return err
			}

			// setup grpc
			grpcServer := components.GRPCServer(ctx.Context, cfg.GRPCServerConfig)
			blockv1.RegisterBlockAPIServer(grpcServer, stores.BlockAPIServer)
			datasetv1.RegisterDatasetAPIServer(grpcServer, stores.DatasetAPIServer)

			// setup api routes
			apiServer := runtime.NewServeMux()
			_ = blockv1.RegisterBlockAPIHandler(ctx.Context, apiServer, serverConn)
			_ = datasetv1.RegisterDatasetAPIHandler(ctx.Context, apiServer, serverConn)

			// prepopulate metrics
			grpc_prometheus.Register(grpcServer)

			// use gin for all other routes (easier to reason about)
			ginServer := components.GinServer(ctx.Context)
			ginServer.Use(
				components.TranslateHeadersToMetadata(),
				func(ginctx *gin.Context) {
					writer := ginctx.Writer
					request := ginctx.Request

					switch {
					case strings.HasPrefix(request.URL.Path, "/v1/fs/"):
						// handle FileServer requests (need to trim prefix)
						fileSystem := &fs.FileSystem{
							Context:    ginctx.Request.Context(),
							BlockAPI:   blockAPI,
							DatasetAPI: datasetAPI,
						}

						handler := http.FileServer(fileSystem)
						handler = http.StripPrefix("/v1/fs/", handler)

						handler.ServeHTTP(writer, request)

					case strings.HasPrefix(request.URL.Path, "/v1/"):
						// handle grpc-gateway requests
						apiServer.ServeHTTP(writer, request)

					}
				},
			)

			err = components.ListenAndServeHTTP(
				ctx.Context,
				cfg.HTTPServerConfig,
				http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					// split grpc here to avoid duplicate prometheus metrics
					if request.ProtoMajor == 2 &&
						strings.HasPrefix(request.Header.Get("Content-Type"), "application/grpc") {
						grpcServer.ServeHTTP(writer, request)
					} else {
						ginServer.ServeHTTP(writer, request)
					}
				}),
			)

			if err != nil {
				return err
			}

			ctxzap.Extract(ctx.Context).Info("running hub")
			<-ctx.Done()
			return nil
		},
		HideHelpCommand: true,
	}
}
