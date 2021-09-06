// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021
package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mjpitz/aetherfs/internal/authors"
	"github.com/mjpitz/aetherfs/internal/commands"
)

//go:embed AUTHORS
var authorsFileContents string

var version string
var commit string
var date string

func main() {
	compiled, err := time.Parse(time.RFC3339, date)
	if err != nil {
		compiled = time.Now()
	}

	if version == "" {
		version = "none"
	}

	if commit == "" {
		commit = "none"
	}

	logLevel := zapcore.InfoLevel
	logFormat := "json"

	app := &cli.App{
		Name:      "aetherfs",
		Usage:     "A publish once, consume many file system for small to medium datasets",
		UsageText: "aetherfs <command>",
		Version:   fmt.Sprintf("%s (%s)", version, commit),
		Commands: []*cli.Command{
			commands.Agent(),
			commands.Login(),
			commands.Logout(),
			commands.Pull(),
			commands.Push(),
			commands.Server(),
		},
		Flags: []cli.Flag{
			&cli.GenericFlag{
				Name:    "log-level",
				Usage:   "the verbosity of logs",
				Value:   &logLevel,
				EnvVars: []string{"LOG_LEVEL"},
			},
			&cli.StringFlag{
				Name:        "log-format",
				Usage:       "how logs should be format",
				Destination: &logFormat,
				Value:       logFormat,
				EnvVars:     []string{"LOG_FORMAT"},
			},
		},
		Before: func(ctx *cli.Context) error {
			cfg := zap.NewProductionConfig()
			cfg.Level.SetLevel(logLevel)
			cfg.Encoding = logFormat

			logger, err := cfg.Build()
			if err != nil {
				return err
			}

			var cancel context.CancelFunc
			ctx.Context, cancel = context.WithCancel(ctx.Context)
			ctx.Context = ctxzap.ToContext(ctx.Context, logger)

			halt := make(chan os.Signal, 1)
			signal.Notify(halt, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-halt
				signal.Stop(halt)

				logger.Info("shutting down")
				cancel()
			}()

			return nil
		},
		After: func(ctx *cli.Context) error {
			// teardown plugins
			//   - destroy outgoing connections
			//   - ensure db files are closed
			//   - etc
			return nil
		},
		Compiled:  compiled,
		Authors:   authors.Parse(authorsFileContents),
		Copyright: fmt.Sprintf("Copyright %d The AetherFS Authors - All Rights Reserved\n", compiled.Year()),
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Print(err)
	}
}
