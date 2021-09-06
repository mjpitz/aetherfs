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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap/zapcore"

	"github.com/mjpitz/aetherfs/internal/authors"
	"github.com/mjpitz/aetherfs/internal/commands"
	"github.com/mjpitz/aetherfs/internal/flagset"
	"github.com/mjpitz/aetherfs/internal/logger"
)

type GlobalConfig struct {
	Log    logger.Config `json:"log,omitempty"`
	State  string        `json:"state,omitempty"  usage:"location where AetherFS can write small amounts of data"`
	Config string        `json:"config,omitempty" usage:"location of the client configuration file"`
}

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
	cfg := &GlobalConfig{
		Log: logger.Config{
			Level:  &logLevel,
			Format: "json",
		},
		State: "/usr/local/aetherfs",
	}

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
		Flags: flagset.Extract(cfg),
		Before: func(ctx *cli.Context) error {
			log, err := logger.Setup(cfg.Log)
			if err != nil {
				return err
			}

			log.Debug("before")

			var cancel context.CancelFunc
			ctx.Context, cancel = context.WithCancel(ctx.Context)
			ctx.Context = ctxzap.ToContext(ctx.Context, log)

			halt := make(chan os.Signal, 1)
			signal.Notify(halt, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-halt
				signal.Stop(halt)

				log.Info("shutting down")
				cancel()
			}()

			return nil
		},
		After: func(ctx *cli.Context) error {
			// teardown plugins
			//   - destroy outgoing connections
			//   - ensure db files are closed
			//   - etc
			ctxzap.Extract(ctx.Context).Debug("after")
			return nil
		},
		Compiled:  compiled,
		Authors:   authors.Parse(authorsFileContents),
		Copyright: fmt.Sprintf("Copyright %d The AetherFS Authors - All Rights Reserved\n", compiled.Year()),
	}

	err = app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
