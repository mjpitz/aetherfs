package commands

import (
	"log"

	"github.com/urfave/cli/v2"
)

func Agent() *cli.Command {
	return &cli.Command{
		Name:        "agent",
		Usage:       "Starts the aetherfs-agent process",
		UsageText:   "aetherfs agent [options]",
		Description: "The aetherfs-agent process is responsible for managing the local file system.",
		Action: func(ctx *cli.Context) error {
			log.Print("running agent")
			<-ctx.Done()
			return nil
		},
	}
}
