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
	"github.com/mjpitz/aetherfs/internal/blocks"
	"github.com/mjpitz/aetherfs/internal/dataset"
	"github.com/mjpitz/aetherfs/internal/storage/local"
	"github.com/mjpitz/myago/flagset"
	"github.com/mjpitz/myago/zaputil"
)

// PushConfig encapsulates all the configuration required to push datasets to AetherFS.
type PushConfig struct {
	BlockSize int32           `json:"block_size"     usage:"the maximum number of bytes per block in MiB"`
	Tags      *dataset.TagSet `json:"tags" alias:"t" usage:"name and tag of the dataset being pushed"`
}

// Push returns a command used to push datasets to upstream servers.
func Push() *cli.Command {
	cfg := &PushConfig{
		BlockSize: 256,
	}

	agentService := &agent.Service{
		Authentications: &local.Authentications{},
	}

	return &cli.Command{
		Name:  "push",
		Usage: "Pushes a dataset into AetherFS",
		UsageText: flagset.ExampleString(
			"aetherfs push [options] <path>",
			"aetherfs push -t maxmind:v1 -t private.company.io/maxmind:v2 /tmp/maxmind",
		),
		Flags: flagset.Extract(cfg),
		Action: func(ctx *cli.Context) error {
			root := ctx.Args().Get(0)
			if root == "" {
				return fmt.Errorf("missing path argument")
			}

			root, err := filepath.Abs(root)
			if err != nil {
				return err
			}

			publishRequest := &agentv1.PublishRequest{
				Sync:      true,
				Path:      root,
				BlockSize: cfg.BlockSize * int32(blocks.Mebibyte),
			}

			for _, tag := range cfg.Tags.Value() {
				publishRequest.Tags = append(publishRequest.Tags, tag.String())
			}

			zaputil.Extract(ctx.Context).Debug("publish", zap.Stringer("request", publishRequest))

			_, err = agentService.Publish(ctx.Context, publishRequest)

			return err
		},
		HideHelpCommand: true,
	}
}
