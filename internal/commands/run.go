// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package commands

import (
	"github.com/urfave/cli/v2"

	"github.com/mjpitz/aetherfs/internal/commands/daemons"
)

// Run returns a command that can execute a given part of the ecosystem.
func Run() *cli.Command {
	return &cli.Command{
		Name:      "run",
		Usage:     "Run the various AetherFS processes",
		UsageText: "aetherfs run <process>",
		Subcommands: []*cli.Command{
			daemons.Agent(),
			daemons.Server(),
		},
	}
}
