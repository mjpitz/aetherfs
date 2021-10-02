// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package commands

import (
	"fmt"
	"os"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/mjpitz/aetherfs/internal/flagset"
)

// PullConfig encapsulates all the configuration required to pull datasets from AetherFS.
type PullConfig struct {
}

// Pull returns a command that downloads datasets from upstream servers
func Pull() *cli.Command {
	cfg := &PullConfig{}

	return &cli.Command{
		Name:  "pull",
		Usage: "Pulls a dataset from AetherFS",
		UsageText: ExampleString(
			"aetherfs pull [options] <path> [dataset...]",
			"aetherfs pull /var/datasets maxmind:v1 private.company.io/maxmind:v2",
			"aetherfs pull -c path/to/application.afs.yaml /var/datasets",
		),
		Flags: flagset.Extract(cfg),
		Action: func(ctx *cli.Context) error {
			logger := ctxzap.Extract(ctx.Context)

			args := ctx.Args().Slice()

			if len(args) == 0 {
				return fmt.Errorf("missing path where we should download datasets")
			}

			path := args[0]
			datasets := args[1:]

			if path == "" {
				path, _ = os.Getwd()
			}

			for _, dataset := range datasets {
				logger.Info("pulling dataset", zap.String("name", dataset), zap.String("path", path))
			}

			return nil
		},
	}
}
