// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	agentv1 "github.com/mjpitz/aetherfs/api/aetherfs/agent/v1"
	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	"github.com/mjpitz/aetherfs/internal/agent"
	"github.com/mjpitz/aetherfs/internal/blocks"
	"github.com/mjpitz/aetherfs/internal/components"
	"github.com/mjpitz/myago/flagset"
)

// PushConfig encapsulates all the configuration required to push datasets to AetherFS.
type PushConfig struct {
	BlockSize int32 `json:"block_size"   usage:"the maximum number of bytes per block in MiB"`
}

// Push returns a command used to push datasets to upstream servers.
func Push() *cli.Command {
	cfg := &PushConfig{
		BlockSize: 256,
	}

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
					Usage:       "name and tag of the dataset being pushed",
					Value:       tags,
					Destination: tags,
					Required:    true,
				},
			}...,
		),
		Action: func(ctx *cli.Context) error {
			root := ctx.Args().Get(0)
			if root == "" {
				return fmt.Errorf("missing path argument")
			}

			root, err := filepath.Abs(root)
			if err != nil {
				return err
			}

			conn, err := components.GRPCClient(ctx.Context, components.GRPCClientConfig{
				Target: lookupHost(),
			})
			if err != nil {
				return err
			}
			defer conn.Close()

			agentService := &agent.Service{
				BlockAPI:   blockv1.NewBlockAPIClient(conn),
				DatasetAPI: datasetv1.NewDatasetAPIClient(conn),
			}

			// cache some metadata for later on to make things easier
			publishRequest := &agentv1.PublishRequest{
				Sync:      true,
				Path:      root,
				BlockSize: cfg.BlockSize * int32(blocks.Mebibyte),
			}

			for _, tag := range tags.Value() {
				parts := strings.Split(tag, ":")
				if len(parts) < 2 {
					parts = append(parts, "latest")
				}

				publishRequest.Tags = append(publishRequest.Tags, &datasetv1.Tag{
					Name:    parts[0],
					Version: parts[1],
				})
			}

			_, err = agentService.Publish(ctx.Context, publishRequest)
			return err
		},
		HideHelpCommand: true,
	}
}
