// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

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
