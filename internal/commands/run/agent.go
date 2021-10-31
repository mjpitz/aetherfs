// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

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

	agentv1 "github.com/mjpitz/aetherfs/api/aetherfs/agent/v1"
	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	"github.com/mjpitz/aetherfs/internal/components"
	"github.com/mjpitz/aetherfs/internal/fs"
	"github.com/mjpitz/myago/flagset"
)

// AgentConfig encapsulates the requirements for configuring and starting up the Agent process.
type AgentConfig struct {
	HTTPServerConfig   components.HTTPServerConfig `json:""`
	GRPCServerConfig   components.GRPCServerConfig `json:""`
	ServerClientConfig components.GRPCClientConfig `json:"server"`
}

// Agent returns a command that will run the agent process.
func Agent() *cli.Command {
	cfg := &AgentConfig{
		HTTPServerConfig: components.HTTPServerConfig{
			Port: 8080,
		},
		ServerClientConfig: components.GRPCClientConfig{
			Target: "aetherfs-hub:8080",
		},
	}

	return &cli.Command{
		Name:        "agent",
		Usage:       "Runs the AetherFS Agent process",
		UsageText:   "aetherfs run agent [options]",
		Description: "The aetherfs-agent process is responsible for managing the local file system.",
		Flags:       flagset.Extract(cfg),
		Action: func(ctx *cli.Context) error {
			agentConn, err := components.GRPCClient(ctx.Context, components.GRPCClientConfig{
				Target:    fmt.Sprintf("localhost:%d", cfg.HTTPServerConfig.Port),
				TLSConfig: cfg.HTTPServerConfig.TLSConfig,
			})
			if err != nil {
				return err
			}

			serverConn, err := components.GRPCClient(ctx.Context, cfg.ServerClientConfig)
			if err != nil {
				return err
			}

			blockAPI := blockv1.NewBlockAPIClient(agentConn)
			datasetAPI := datasetv1.NewDatasetAPIClient(serverConn)

			agentSvc := &agentService{
				blockAPI:   blockAPI,
				datasetAPI: datasetAPI,
			}
			blockSvc := &blockv1.UnimplementedBlockAPIServer{}

			// setup grpc
			grpcServer := components.GRPCServer(ctx.Context, cfg.GRPCServerConfig)
			agentv1.RegisterAgentAPIServer(grpcServer, agentSvc)
			blockv1.RegisterBlockAPIServer(grpcServer, blockSvc)

			// setup api routes
			apiServer := runtime.NewServeMux()
			_ = agentv1.RegisterAgentAPIHandler(ctx.Context, apiServer, agentConn)
			_ = blockv1.RegisterBlockAPIHandler(ctx.Context, apiServer, agentConn)

			// prepopulate metrics
			grpc_prometheus.Register(grpcServer)

			// use gin for all other routes (easier to reason about)
			ginServer := components.GinServer(ctx.Context)
			ginServer.Use(components.TranslateHeadersToMetadata())

			ginServer.Group("/api").Any("*path", gin.WrapH(apiServer))
			ginServer.Group("/fs").GET("*path", func(ginctx *gin.Context) {
				// handle FileServer requests (need to trim prefix)
				fileSystem := &fs.FileSystem{
					Context:    ginctx.Request.Context(),
					BlockAPI:   blockAPI,
					DatasetAPI: datasetAPI,
				}

				handler := http.FileServer(fileSystem)
				handler = http.StripPrefix("/fs/", handler)

				handler.ServeHTTP(ginctx.Writer, ginctx.Request)
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

			ctxzap.Extract(ctx.Context).Info("running agent")
			<-ctx.Done()
			return nil
		},
		HideHelpCommand: true,
		Hidden:          true,
	}
}
