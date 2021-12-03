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

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mjpitz/aetherfs/internal/commands"
	"github.com/mjpitz/myago/authors"
	"github.com/mjpitz/myago/flagset"
	"github.com/mjpitz/myago/lifecycle"
	"github.com/mjpitz/myago/zaputil"
)

//go:embed AUTHORS
var authorsFileContents string

var version = "none"
var commit = "none"
var date = time.Now().Format(time.RFC3339)

type GlobalConfig struct {
	Log zaputil.Config `json:"log"`
	//StateDir string        `json:"state_dir" usage:"location where AetherFS can write small amounts of data"`
	//Config   string        `json:"config"    usage:"location of the command configuration file"`
}

func main() {
	compiled, _ := time.Parse(time.RFC3339, date)

	cfg := &GlobalConfig{
		Log: zaputil.DefaultConfig(),
		//StateDir: "/usr/local/aetherfs",
	}

	app := &cli.App{
		Name:      "aetherfs",
		Usage:     "A virtual file system for small to medium sized datasets (MB or GB, not TB or PB).",
		UsageText: "aetherfs [options] <command>",
		Version:   fmt.Sprintf("%s (%s)", version, commit),
		Commands:  commands.Available,
		Flags:     flagset.Extract(cfg),
		Before: func(ctx *cli.Context) error {
			ctx.Context = zaputil.Setup(ctx.Context, cfg.Log)
			ctx.Context = lifecycle.Setup(ctx.Context)

			// special grpc things
			logger := zaputil.Extract(ctx.Context)
			ctx.Context = ctxzap.ToContext(ctx.Context, logger)

			if cfg.Log.Level != zapcore.DebugLevel.String() {
				logger = zap.NewNop()
			}
			grpc_zap.ReplaceGrpcLoggerV2(logger)

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
