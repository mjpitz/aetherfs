package commands

import (
	"log"

	"github.com/urfave/cli/v2"

	"github.com/mjpitz/aetherfs/internal/auth"
)

// ServerConfig encapsulates the requirements for configuring and starting up the Server process.
type ServerConfig struct {
	OIDCIssuer auth.OIDCIssuer `json:"oidc_issuer,omitempty"`
}

// Server returns a cli.Command that can be added to an existing application.
func Server() *cli.Command {
	return &cli.Command{
		Name:        "server",
		Usage:       "Starts the AetherFS Server process",
		UsageText:   "aetherfs server [options]",
		Description: "The aetherfs-server process is responsible for the datasets in our small blob store.",
		Action: func(ctx *cli.Context) error {
			log.Print("running server")
			<-ctx.Done()
			return nil
		},
	}
}
