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
	cli "github.com/urfave/cli/v2"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	fsv1 "github.com/mjpitz/aetherfs/api/aetherfs/fs/v1"
	"github.com/mjpitz/aetherfs/internal/components"
	"github.com/mjpitz/aetherfs/internal/flagset"
)

// ServerConfig encapsulates the requirements for configuring and starting up the Server process.
type ServerConfig struct {
	GRPCServerConfig components.GRPCServerConfig `json:",omitempty"`
	HTTPServerConfig components.HTTPServerConfig `json:",omitempty"`
}

// Server returns a cli.Command that can be added to an existing application.
func Server() *cli.Command {
	cfg := &ServerConfig{
		HTTPServerConfig: components.HTTPServerConfig{
			Port: 8080,
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

			blockSvc := &blockService{}
			datasetSvc := &datasetService{}
			fileServerSvc := &fileServerService{}

			// setup grpc
			grpcServer := components.GRPCServer(ctx.Context, cfg.GRPCServerConfig)
			blockv1.RegisterBlockAPIServer(grpcServer, blockSvc)
			datasetv1.RegisterDatasetAPIServer(grpcServer, datasetSvc)
			fsv1.RegisterFileServerAPIServer(grpcServer, fileServerSvc)

			// setup api routes
			apiServer := runtime.NewServeMux()
			_ = blockv1.RegisterBlockAPIHandler(ctx.Context, apiServer, serverConn)
			_ = datasetv1.RegisterDatasetAPIHandler(ctx.Context, apiServer, serverConn)
			_ = fsv1.RegisterFileServerAPIHandler(ctx.Context, apiServer, serverConn)

			// prepopulate metrics
			grpc_prometheus.Register(grpcServer)

			// use gin for all other routes (easier to reason about)
			ginServer := components.GinServer(ctx.Context)
			ginServer.Use(func(ginctx *gin.Context) {
				if strings.HasPrefix(ginctx.Request.URL.Path, "/v1/") {
					apiServer.ServeHTTP(ginctx.Writer, ginctx.Request)
				}
			})

			err = components.ListenAndServeHTTP(
				ctx.Context,
				cfg.HTTPServerConfig,
				http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
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
	}
}
