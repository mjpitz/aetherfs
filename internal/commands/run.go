// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package commands

import (
	"github.com/urfave/cli/v2"

	"github.com/mjpitz/aetherfs/internal/commands/daemons"
	"github.com/mjpitz/aetherfs/internal/flagset"
)

// RunConfig encapsulates all the configuration required to start an AetherFS process.
type RunConfig struct{}

// Start returns a cli.Command that can be added to an existing application.
func Run() *cli.Command {
	cfg := &RunConfig{}

	return &cli.Command{
		Name:      "run",
		Usage:     "Run the various AetherFS processes",
		UsageText: "aetherfs run <process>",
		Flags:     flagset.Extract(cfg),
		Subcommands: []*cli.Command{
			daemons.Agent(),
			daemons.Server(),
		},
	}
}
