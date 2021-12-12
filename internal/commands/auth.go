// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package commands

import (
	"fmt"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/mjpitz/aetherfs/internal/components"
	"github.com/mjpitz/aetherfs/internal/storage/local"
	"github.com/mjpitz/myago/flagset"
)

// AuthConfig encapsulates all the configuration required to log in to an AetherFS instance.
type AuthConfig struct {
	components.GRPCClientConfig
}

// Auth returns a cli.Command that can be added to an existing application.
func Auth() *cli.Command {
	cfg := &AuthConfig{}
	authService := &local.Authentications{}

	return &cli.Command{
		Name:      "auth",
		Usage:     "Manage authentication to AetherFS servers",
		UsageText: "aetherfs auth <command>",
		Subcommands: []*cli.Command{
			{
				Name:      "add",
				Usage:     "Adds authentication to an AetherFS server",
				UsageText: "aetherfs auth add [options] <server>",
				Flags:     flagset.Extract(cfg),
				Action: func(ctx *cli.Context) error {
					server := ctx.Args().Get(0)
					if len(server) == 0 {
						return fmt.Errorf("server name not provided")
					}

					if len(cfg.Target) == 0 {
						cfg.Target = server
					}

					return authService.Put(ctx.Context, &local.Credentials{
						Server:           server,
						GRPCClientConfig: cfg.GRPCClientConfig,
					})
				},
				HideHelpCommand: true,
			},
			{
				Name:      "remove",
				Usage:     "Removes authentication to an AetherFS server",
				UsageText: "aetherfs auth remove <server>",
				Action: func(ctx *cli.Context) error {
					server := ctx.Args().Get(0)
					if len(server) == 0 {
						return fmt.Errorf("server name not provided")
					}

					return authService.Delete(ctx.Context, server)
				},
				HideHelpCommand: true,
			},
			{
				Name:      "show",
				Usage:     "Shows non-sensitive authentication information for an AetherFS server",
				UsageText: "aetherfs auth show <server>",
				Action: func(ctx *cli.Context) error {
					server := ctx.Args().Get(0)
					if len(server) == 0 {
						return fmt.Errorf("server name not provided")
					}

					creds, err := authService.Get(ctx.Context, server)
					if err != nil {
						return err
					}

					t, err := template.New("details").Parse(details)
					if err != nil {
						return err
					}

					return t.Execute(ctx.App.Writer, creds)
				},
				HideHelpCommand: true,
			},
		},
		HideHelpCommand: true,
	}
}

const details = `
SERVER: {{ .Server }}
TARGET: {{ .GRPCClientConfig.Target }}
{{ if .GRPCClientConfig.TLS.Enable }}
TLS
===
{{- if .GRPCClientConfig.TLS.CertPath }}
CERT PATH: {{ .GRPCClientConfig.TLS.CertPath }}
CA FILE:   {{ .GRPCClientConfig.TLS.CAFile }}
CERT FILE: {{ .GRPCClientConfig.TLS.CertFile }}
KEY FILE:  {{ .GRPCClientConfig.TLS.KeyFile }}
{{- else }}
CERT PATH: <system>
{{- end }}
{{ end }}
`
