package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func Pull() *cli.Command {

	return &cli.Command{
		Name:  "pull",
		Usage: "Pulls a dataset from AetherFS",
		UsageText: ExampleString(
			"aetherfs pull [options] <path> [dataset...]",
			"aetherfs pull /var/datasets maxmind:v1 private.company.io/maxmind:v2",
			"aetherfs pull -c path/to/application.afs.yaml /var/datasets",
		),
		Action: func(ctx *cli.Context) error {
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
				log.Printf("pulling dataset %s", dataset)
			}

			return nil
		},
	}
}
