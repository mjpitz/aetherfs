// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package commands

import (
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/mjpitz/aetherfs/internal/flagset"
)

// PushConfig encapsulates all the configuration required to push datasets to AetherFS.
type PushConfig struct {
}

// Push returns a command used to push datasets to upstream servers.
func Push() *cli.Command {
	cfg := &PushConfig{}
	tags := cli.NewStringSlice() // can't put this in config struct quite yet

	return &cli.Command{
		Name:  "push",
		Usage: "Pushes a dataset into AetherFS",
		UsageText: ExampleString(
			"aetherfs push [options] <path>",
			"aetherfs push -t maxmind:v1 -t private.company.io/maxmind:v2 /tmp/maxmind",
		),
		Flags: append(
			flagset.Extract(cfg),
			[]cli.Flag{
				&cli.StringSliceFlag{
					Name:        "tag",
					Aliases:     []string{"t"},
					Usage:       "name and tag of the dataset we're pushing",
					Value:       tags,
					Destination: tags,
					Required:    true,
				},
			}...,
		),
		Action: func(ctx *cli.Context) error {
			logger := ctxzap.Extract(ctx.Context)

			path := ctx.Args().Get(0)

			if path == "" {
				return fmt.Errorf("missing argument: path")
			}

			for _, tag := range tags.Value() {
				logger.Info("pushing dataset", zap.String("name", tag), zap.String("path", path))
			}

			return nil
		},
	}
}
