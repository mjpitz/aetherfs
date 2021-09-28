// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021
package main

import (
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap/zapcore"

	"github.com/mjpitz/aetherfs/internal/authors"
	"github.com/mjpitz/aetherfs/internal/commands"
	"github.com/mjpitz/aetherfs/internal/flagset"
	"github.com/mjpitz/aetherfs/internal/lifecycle"
	"github.com/mjpitz/aetherfs/internal/logger"
)

//go:embed AUTHORS
var authorsFileContents string

var version = "none"
var commit = "none"
var date = time.Now().Format(time.RFC3339)

type GlobalConfig struct {
	Log      logger.Config `json:"log,omitempty"`
	StateDir string        `json:"state_dir,omitempty" usage:"location where AetherFS can write small amounts of data"`
	Config   string        `json:"config,omitempty"    usage:"location of the command configuration file"`
}

func main() {
	compiled, _ := time.Parse(time.RFC3339, date)

	format := "json"
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		format = "console" // looks like terminal session, use console logging
	}

	logLevel := zapcore.InfoLevel
	cfg := &GlobalConfig{
		Log: logger.Config{
			Level:  &logLevel,
			Format: format,
		},
		StateDir: "/usr/local/aetherfs",
	}

	app := &cli.App{
		Name:      "aetherfs",
		Usage:     "A publish once, consume many file system for small to medium datasets",
		UsageText: "aetherfs [options] <command>",
		Version:   fmt.Sprintf("%s (%s)", version, commit),
		Commands: []*cli.Command{
			commands.Auth(),
			commands.Pull(),
			commands.Push(),
			commands.Run(),
		},
		Flags: flagset.Extract(cfg),
		Before: func(ctx *cli.Context) error {
			ctx.Context = logger.Setup(ctx.Context, cfg.Log)
			ctx.Context = lifecycle.Setup(ctx.Context)

			return nil
		},
		After: func(ctx *cli.Context) error {
			lifecycle.Resolve(ctx.Context)

			return nil
		},
		Compiled:             compiled,
		Authors:              authors.Parse(authorsFileContents),
		Copyright:            fmt.Sprintf("Copyright %d The AetherFS Authors - All Rights Reserved\n", compiled.Year()),
		HideHelpCommand:      true,
		EnableBashCompletion: true,
		BashComplete:         cli.DefaultAppComplete,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
