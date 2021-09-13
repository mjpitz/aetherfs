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

	agentv1 "github.com/mjpitz/aetherfs/api/aetherfs/agent/v1"
	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	fsv1 "github.com/mjpitz/aetherfs/api/aetherfs/fs/v1"
	"github.com/mjpitz/aetherfs/internal/components"
	"github.com/mjpitz/aetherfs/internal/flagset"
)

// AgentConfig encapsulates the requirements for configuring and starting up the Agent process.
type AgentConfig struct {
	GRPCServerConfig components.GRPCServerConfig `json:",omitempty"`
	HTTPServerConfig components.HTTPServerConfig `json:",omitempty"`

	ServerClientConfig components.GRPCClientConfig `json:"server,omitempty"`
}

// Agent returns a cli.Command that can be added to an existing application.
func Agent() *cli.Command {
	cfg := &AgentConfig{
		HTTPServerConfig: components.HTTPServerConfig{
			Port: 8080,
		},
		ServerClientConfig: components.GRPCClientConfig{
			Target: "aetherfs-server:8080",
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

			agentSvc := &agentService{
				blockAPI:   blockv1.NewBlockAPIClient(agentConn),
				datasetAPI: datasetv1.NewDatasetAPIClient(serverConn),
			}
			blockSvc := &blockService{}
			fileServerSvc := &fileServerService{}

			// setup grpc
			grpcServer := components.GRPCServer(ctx.Context, cfg.GRPCServerConfig)
			agentv1.RegisterAgentAPIServer(grpcServer, agentSvc)
			blockv1.RegisterBlockAPIServer(grpcServer, blockSvc)
			fsv1.RegisterFileServerAPIServer(grpcServer, fileServerSvc)

			// setup api routes
			apiServer := runtime.NewServeMux()
			_ = agentv1.RegisterAgentAPIHandler(ctx.Context, apiServer, agentConn)
			_ = blockv1.RegisterBlockAPIHandler(ctx.Context, apiServer, agentConn)
			_ = fsv1.RegisterFileServerAPIHandler(ctx.Context, apiServer, agentConn)

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

			ctxzap.Extract(ctx.Context).Info("running agent")
			<-ctx.Done()
			return nil
		},
	}
}
