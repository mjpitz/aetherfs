// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package commands

import (
	"github.com/urfave/cli/v2"

	"github.com/mjpitz/aetherfs/internal/commands/run"
)

// Run returns a command that can execute a given part of the ecosystem.
func Run() *cli.Command {
	return &cli.Command{
		Name:      "run",
		Usage:     "Run the various AetherFS processes",
		UsageText: "aetherfs run <process>",
		Subcommands: []*cli.Command{
			//daemons.Agent(),
			run.Hub(),
		},
		HideHelpCommand: true,
	}
}
