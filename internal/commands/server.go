package commands

import (
	"log"

	"github.com/urfave/cli/v2"
)

func Server() *cli.Command {
	return &cli.Command{
		Name:        "server",
		Usage:       "Starts the aetherfs-server process",
		UsageText:   "aetherfs server [options]",
		Description: "The aetherfs-server process is responsible for the datasets in our small blob store.",
		Action: func(ctx *cli.Context) error {
			log.Print("running server")
			<-ctx.Done()
			return nil
		},
	}
}
