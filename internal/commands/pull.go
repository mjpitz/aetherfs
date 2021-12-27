// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package commands

import (
	"fmt"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	agentv1 "github.com/mjpitz/aetherfs/api/aetherfs/agent/v1"
	"github.com/mjpitz/aetherfs/internal/agent"
	"github.com/mjpitz/aetherfs/internal/storage/local"
	"github.com/mjpitz/myago/flagset"
	"github.com/mjpitz/myago/zaputil"
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
		UsageText: flagset.ExampleString(
			"aetherfs pull [options] <path> [dataset...]",
			"aetherfs pull /var/datasets maxmind:v1 private.company.io/maxmind:v2",
			"aetherfs pull -c path/to/application.afs.yaml /var/datasets",
		),
		Flags: flagset.Extract(cfg),
		Action: func(ctx *cli.Context) error {
			args := ctx.Args().Slice()
			switch len(args) {
			case 0:
				return fmt.Errorf("missing required path")
			case 1:
				return fmt.Errorf("missing datasets")
			}

			root, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			subscribeRequest := &agentv1.SubscribeRequest{
				Sync: true,
				Path: root,
				Tags: args[1:],
			}

			zaputil.Extract(ctx.Context).Debug("subscribe", zap.Stringer("request", subscribeRequest))

			agentService := &agent.Service{
				Credentials: local.Extract(ctx.Context).Credentials(),
			}

			_, err = agentService.Subscribe(ctx.Context, subscribeRequest)

			return err
		},
		HideHelpCommand: true,
	}
}
