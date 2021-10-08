// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package main

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

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
	Log logger.Config `json:"log"`
	//StateDir string        `json:"state_dir" usage:"location where AetherFS can write small amounts of data"`
	//Config   string        `json:"config"    usage:"location of the command configuration file"`
}

func main() {
	compiled, _ := time.Parse(time.RFC3339, date)

	format := "json"
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		format = "console" // looks like terminal session, use console logging
	}

	cfg := &GlobalConfig{
		Log: logger.Config{
			Level:  "info",
			Format: format,
		},
		//StateDir: "/usr/local/aetherfs",
	}

	app := &cli.App{
		Name:      "aetherfs",
		Usage:     "A virtual file system for small to medium sized datasets (MB or GB, not TB or PB).",
		UsageText: "aetherfs [options] <command>",
		Version:   fmt.Sprintf("%s (%s)", version, commit),
		Commands: []*cli.Command{
			//commands.Auth(),
			commands.Pull(),
			commands.Push(),
			commands.Run(),
			commands.Version(),
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
		HideVersion:          true,
		HideHelpCommand:      true,
		EnableBashCompletion: true,
		BashComplete:         cli.DefaultAppComplete,
		Metadata: map[string]interface{}{
			"arch":       runtime.GOARCH,
			"compiled":   date,
			"go_version": strings.TrimPrefix(runtime.Version(), "go"),
			"os":         runtime.GOOS,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
