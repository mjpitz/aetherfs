// AetherFS - A virtual file system for small to medium sized datasets (MB or GB, not TB or PB).
// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package commands

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/urfave/cli/v2"

	"github.com/mjpitz/aetherfs/internal/auth"
	"github.com/mjpitz/aetherfs/internal/components"
	"github.com/mjpitz/aetherfs/internal/flagset"
)

// AuthConfig encapsulates all the configuration required to log in to an AetherFS instance.
type AuthConfig struct {
	auth.Config                 `json:""`
	components.GRPCClientConfig `json:""`

	Remove bool `json:"remove" usage:"set to prune existing credentials"`
}

// Auth returns a cli.Command that can be added to an existing application.
func Auth() *cli.Command {
	cfg := &AuthConfig{}

	return &cli.Command{
		Name:        "authenticate",
		Usage:       "Manage authentication to an AetherFS instance",
		UsageText:   "aetherfs auth [options] <server_url>",
		Description: "",
		Flags:       flagset.Extract(cfg),
		Aliases:     []string{"auth"},
		Action: func(ctx *cli.Context) error {
			if cfg.Remove {
				ctxzap.Extract(ctx.Context).Info("logging out")
			} else {
				ctxzap.Extract(ctx.Context).Info("logging in")
			}
			return nil
		},
		HideHelpCommand: true,
	}
}
