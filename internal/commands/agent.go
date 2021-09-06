package commands

import (
	"log"

	"github.com/urfave/cli/v2"

	"github.com/mjpitz/aetherfs/internal/auth"
	"github.com/mjpitz/aetherfs/internal/flagset"
)

// AgentConfig encapsulates the requirements for configuring and starting up the Agent process.
type AgentConfig struct {
	OIDC struct {
		Issuer auth.OIDCIssuer `json:"issuer,omitempty"`
	} `json:"oidc,omitempty"`
}

// Agent returns a cli.Command that can be added to an existing application.
func Agent() *cli.Command {
	cfg := &AgentConfig{}

	return &cli.Command{
		Name:        "agent",
		Usage:       "Starts the AetherFS Agent process",
		UsageText:   "aetherfs agent [options]",
		Description: "The aetherfs-agent process is responsible for managing the local file system.",
		Flags:       flagset.Extract(cfg),
		Action: func(ctx *cli.Context) error {
			log.Print("running agent")
			<-ctx.Done()
			return nil
		},
	}
}
