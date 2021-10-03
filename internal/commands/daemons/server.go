// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package daemons

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc/metadata"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	"github.com/mjpitz/aetherfs/internal/components"
	"github.com/mjpitz/aetherfs/internal/flagset"
	"github.com/mjpitz/aetherfs/internal/fs"
	"github.com/mjpitz/aetherfs/internal/storage"
	"github.com/mjpitz/aetherfs/internal/storage/s3"
)

// ServerConfig encapsulates the requirements for configuring and starting up the Server process.
type ServerConfig struct {
	HTTPServerConfig components.HTTPServerConfig `json:""`
	GRPCServerConfig components.GRPCServerConfig `json:""`
	StorageConfig    storage.Config              `json:"storage"`
}

// Server returns a command that will run the server process.
func Server() *cli.Command {
	cfg := &ServerConfig{
		HTTPServerConfig: components.HTTPServerConfig{
			Port: 8080,
		},
		StorageConfig: storage.Config{
			Driver: "s3",
			S3: s3.Config{
				Endpoint: "s3.amazonaws.com",
				TLS: components.TLSConfig{
					Enable: true,
				},
			},
		},
	}

	return &cli.Command{
		Name:        "server",
		Usage:       "Runs the AetherFS Server process",
		UsageText:   "aetherfs run server [options]",
		Description: "The aetherfs-server process is responsible for the datasets in our small blob store.",
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
			ginServer.Use(func(ginctx *gin.Context) {
				// preprocess headers into grpc metadata
				md := metadata.New(nil)
				for k, vv := range ginctx.Request.Header {
					md.Set(k, vv...)
				}

				ctx := metadata.NewIncomingContext(ginctx.Request.Context(), md)
				ginctx.Request = ginctx.Request.WithContext(ctx)

				writer := ginctx.Writer
				request := ginctx.Request

				switch {
				case strings.HasPrefix(request.URL.Path, "/v1/fs/"):
					// handle FileServer requests (need to trim prefix)
					request.URL.Path = strings.TrimPrefix(request.URL.Path, "/v1/fs/")

					fileSystem := &fs.FileSystem{
						Context:    ctx,
						BlockAPI:   blockAPI,
						DatasetAPI: datasetAPI,
					}

					http.FileServer(fileSystem).ServeHTTP(writer, request)

				case strings.HasPrefix(request.URL.Path, "/v1/"):
					// handle grpc-gateway requests
					apiServer.ServeHTTP(writer, request)

				}
			})

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

			ctxzap.Extract(ctx.Context).Info("running server")
			<-ctx.Done()
			return nil
		},
		HideHelpCommand: true,
	}
}
