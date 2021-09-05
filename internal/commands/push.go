package commands

import (
	"fmt"
	"log"

	"github.com/urfave/cli/v2"
)

func Push() *cli.Command {
	tags := cli.NewStringSlice()

	return &cli.Command{
		Name:  "push",
		Usage: "Pushes a dataset into AetherFS",
		UsageText: ExampleString(
			"aetherfs push [options] <path>",
			"aetherfs push -t maxmind:v1 -t private.company.io/maxmind:v2 /tmp/maxmind",
		),
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "tag",
				Aliases:     []string{"t"},
				Usage:       "name and optional tag to use for the dataset",
				Destination: tags,
				Value:       tags,
				EnvVars:     []string{},
			},
		},
		Action: func(ctx *cli.Context) error {
			path := ctx.Args().Get(0)

			if path == "" {
				return fmt.Errorf("missing argument: path")
			}

			if len(tags.Value()) == 0 {
				return fmt.Errorf("missing option: tag - at least one must be provided")
			}

			for _, tag := range tags.Value() {
				log.Printf("pushing dataset %s", tag)
			}

			return nil
		},
	}
}
