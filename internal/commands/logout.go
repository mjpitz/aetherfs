// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package commands

import (
	"github.com/urfave/cli/v2"

	"github.com/mjpitz/aetherfs/internal/flagset"
)

// LogoutConfig encapsulates all the configuration required to logout of an AetherFS instance.
type LogoutConfig struct{}

// Logout returns a cli.Command that can be added to an existing application.
func Logout() *cli.Command {
	cfg := &LogoutConfig{}

	return &cli.Command{
		Name:        "logout",
		Usage:       "Log out of an AetherFS instance",
		UsageText:   "aetherfs logout <server_url>",
		Description: "",
		Flags:       flagset.Extract(cfg),
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}
