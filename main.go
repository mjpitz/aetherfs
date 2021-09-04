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

	"github.com/urfave/cli/v2"

	"github.com/mjpitz/aetherfs/internal/authors"
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
		version = "dev"
	}

	if commit == "" {
		commit = "HEAD"
	}

	app := &cli.App{
		Name:        "aetherfs",
		Usage:       "A publish once, consume many file system for small to medium datasets",
		UsageText:   "aetherfs <command>",
		Version:     fmt.Sprintf("%s (%s)", version, commit),
		Commands: []*cli.Command{
			{
				Name:      "agent",
				Usage:     "Starts the aetherfs-agent process",
				UsageText: "aetherfs agent [options]",
				Action: func(ctx *cli.Context) error {
					log.Print("running agent")
					<-ctx.Done()
					return nil
				},
			},
			{
				Name:      "server",
				Usage:     "Starts the aetherfs-server process",
				UsageText: "aetherfs server [options]",
				Action: func(ctx *cli.Context) error {
					log.Print("running server")
					<-ctx.Done()
					return nil
				},
			},
			{
				Name:      "push",
				Usage:     "Pushes a dataset into AetherFS",
				UsageText: "aetherfs push [options] <path>",
				Action: func(ctx *cli.Context) error {
					log.Print("pushing dataset")
					return nil
				},
			},
			{
				Name:      "pull",
				Usage:     "Pulls a dataset from AetherFS",
				UsageText: "aetherfs pull [options] <dataset> [path]",
				Action: func(ctx *cli.Context) error {
					log.Print("pulling dataset")
					return nil
				},
			},
		},
		Flags: []cli.Flag{},
		Before: func(ctx *cli.Context) error {
			var cancel context.CancelFunc
			ctx.Context, cancel = context.WithCancel(ctx.Context)

			halt := make(chan os.Signal, 1)
			signal.Notify(halt, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-halt
				signal.Stop(halt)

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
		Copyright: fmt.Sprintf("Copyright %d The AetherFS Authors - All Rights Reserved", compiled.Year()),
	}

	_ = app.Run(os.Args)
}
