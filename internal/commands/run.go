// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package commands

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	agentv1 "github.com/mjpitz/aetherfs/api/aetherfs/agent/v1"
	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	"github.com/mjpitz/aetherfs/internal/afs"
	"github.com/mjpitz/aetherfs/internal/agent"
	"github.com/mjpitz/aetherfs/internal/components"
	"github.com/mjpitz/aetherfs/internal/storage"
	"github.com/mjpitz/aetherfs/internal/storage/local"
	"github.com/mjpitz/aetherfs/internal/web"
	"github.com/mjpitz/myago/config"
	"github.com/mjpitz/myago/flagset"
	"github.com/mjpitz/myago/zaputil"
)

type RunConfig struct {
	ConfigFile string `json:"config_file" usage:"specify the location of a file containing the run configuration"`

	components.HTTPServerConfig
	components.GRPCServerConfig

	NFS     components.NFSServerConfig `json:"nfs"`
	Agent   agent.Config               `json:"agent"`
	Storage storage.Config             `json:"storage"`
	Web     web.Config                 `json:"web"`
}

// Run returns a command that can execute a given part of the ecosystem.
func Run() (cmd *cli.Command) {
	cfg := &RunConfig{}

	cmd = &cli.Command{
		Name:      "run",
		Usage:     "Runs the AetherFS process",
		UsageText: "aetherfs run [options]",
		Flags:     flagset.Extract(cfg),
		Before: func(ctx *cli.Context) error {
			if cfg.ConfigFile != "" {
				err := config.Load(ctx.Context, cfg, cfg.ConfigFile)
				if err != nil {
					return err
				}
			}

			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:            "hub",
				Usage:           "Runs the AetherFS Hub process",
				UsageText:       "aetherfs run hub [options]",
				Description:     "The aetherfs-hub process is responsible for collecting and hosting datasets.",
				Flags:           flagset.Extract(cfg),
				Hidden:          true,
				HideHelpCommand: true,
			},
		},
		Action: func(ctx *cli.Context) error {
			log := zaputil.Extract(ctx.Context)

			serverConn, err := components.GRPCClient(ctx.Context, components.GRPCClientConfig{
				Target: fmt.Sprintf("localhost:%d", cfg.HTTPServerConfig.Port),
				TLS:    cfg.HTTPServerConfig.TLS,
			})
			if err != nil {
				return err
			}

			blockAPI := blockv1.NewBlockAPIClient(serverConn)
			datasetAPI := datasetv1.NewDatasetAPIClient(serverConn)

			stores, err := storage.ObtainStores(ctx.Context, cfg.Storage)
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

			if cfg.Agent.Enable {

				log.Info("enabling", zap.Strings("components", []string{"agent"}))
				agentService := &agent.Service{
					Credentials: local.Extract(ctx.Context).Credentials(),
				}

				if cfg.Agent.Shutdown.Enable {
					log.Info("enabling shutdown", zap.Strings("components", []string{"agent"}))
					ctx.Context, agentService.InitiateShutdown = context.WithCancel(ctx.Context)
				}

				agentv1.RegisterAgentAPIServer(grpcServer, agentService)
				_ = agentv1.RegisterAgentAPIHandler(ctx.Context, apiServer, serverConn)
			}

			// prepopulate metrics
			grpc_prometheus.Register(grpcServer)

			// use gin for all other routes (easier to reason about)
			ginServer := components.GinServer(ctx.Context)
			ginServer.Use(components.TranslateHeadersToMetadata())

			ginServer.Group("/api").Any("*path", gin.WrapH(apiServer))

			ginServer.Group("/fs").GET("*path", func(ginctx *gin.Context) {
				// handle FileServer requests (need to trim prefix)
				fileSystem := afero.NewHttpFs(&afs.FileSystem{
					Context:    ginctx.Request.Context(),
					BlockAPI:   blockAPI,
					DatasetAPI: datasetAPI,
				})

				handler := http.FileServer(fileSystem)
				handler = http.StripPrefix("/fs/", handler)

				handler.ServeHTTP(ginctx.Writer, ginctx.Request)
			})

			if cfg.Web.Enable {
				log.Info("enabling", zap.Strings("components", []string{"web"}))

				ginServer.Group("/ui").GET("*path", gin.WrapH(web.Handle()))

				ginServer.GET("/", func(ginctx *gin.Context) {
					ginctx.Redirect(http.StatusTemporaryRedirect, "/ui/")
				})
			}

			if cfg.NFS.Enable {
				log.Info("enabling",
					zap.Strings("components", []string{"nfs"}),
					zap.Int("port", cfg.NFS.Port))

				err = components.ListenAndServeNFS(ctx.Context, cfg.NFS, &afs.FileSystem{
					Context:    ctx.Context,
					BlockAPI:   blockAPI,
					DatasetAPI: datasetAPI,
				})
				if err != nil {
					return err
				}
			}

			log.Info("enabling",
				zap.Strings("components", []string{"http", "grpc"}),
				zap.Int("port", cfg.HTTPServerConfig.Port))

			err = components.ListenAndServeHTTP(
				ctx.Context,
				cfg.HTTPServerConfig,
				http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					// split grpc here to avoid duplicate prometheus metrics
					switch {
					case request.ProtoMajor == 2 &&
						strings.HasPrefix(request.Header.Get("Content-Type"), "application/grpc"):
						grpcServer.ServeHTTP(writer, request)

					case strings.HasPrefix(request.URL.Path, "/webdav"):
						// webdav has to be done here for custom HTTP methods
						fileSystem := &afs.FileSystem{
							Context:    request.Context(),
							BlockAPI:   blockAPI,
							DatasetAPI: datasetAPI,
						}

						handler := afs.Webdav(fileSystem)
						handler.Prefix = "/webdav"

						handler.ServeHTTP(writer, request)
					default:
						ginServer.ServeHTTP(writer, request)
					}
				}),
			)

			if err != nil {
				return err
			}

			log.Info("running aetherfs")
			<-ctx.Done()
			return nil
		},
		HideHelpCommand: true,
	}

	// for backward compatibility
	cmd.Subcommands[0].Action = cmd.Action

	return cmd
}
