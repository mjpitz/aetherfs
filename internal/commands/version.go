// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package commands

import (
	"text/template"

	"github.com/urfave/cli/v2"
)

const versionTemplate = "{{ .Name }} {{ .Version }} {{ range $key, $value := .Metadata }}{{ $key }}={{ $value }} {{ end }}\n"

// Version returns a command that outputs version information for the application.
func Version() *cli.Command {
	return &cli.Command{
		Name:      "version",
		Usage:     "Print the binary version information",
		UsageText: "aetherfs version",
		Action: func(ctx *cli.Context) error {
			return template.
				Must(template.New("version").Parse(versionTemplate)).
				Execute(ctx.App.Writer, ctx.App)
		},
		HideHelpCommand: true,
	}
}
