package commands

import (
	"github.com/urfave/cli/v2"

	"github.com/mjpitz/aetherfs/internal/auth"
	"github.com/mjpitz/aetherfs/internal/flagset"
)

// LoginConfig encapsulates all the configuration required to log in to an AetherFS instance.
type LoginConfig struct {
	auth.Config `json:",omitempty"`
}

// Login returns a cli.Command that can be added to an existing application.
func Login() *cli.Command {
	cfg := &LoginConfig{}

	return &cli.Command{
		Name:        "login",
		Usage:       "Log in to an AetherFS instance",
		UsageText:   "aetherfs login [options] <server_url>",
		Description: "",
		Flags:       flagset.Extract(cfg),
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}
