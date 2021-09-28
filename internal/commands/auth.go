// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package commands

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/urfave/cli/v2"

	"github.com/mjpitz/aetherfs/internal/auth"
	"github.com/mjpitz/aetherfs/internal/flagset"
)

// AuthConfig encapsulates all the configuration required to log in to an AetherFS instance.
type AuthConfig struct {
	auth.Config `json:",omitempty"`

	Remove bool `json:"remove,omitempty" usage:"set to prune existing credentials"`
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
	}
}
