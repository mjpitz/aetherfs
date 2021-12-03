// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

//go:build linux
// +build linux

package commands

import (
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
)

func init() {
	cmd := VFS()

	idx := sort.Search(len(Available), func(i int) bool {
		return strings.Compare(cmd.Name, Available[i].Name) < 0
	})

	if idx == len(Available) {
		Available = append(Available, cmd)
	} else {
		Available = append(Available[:idx], append([]*cli.Command{cmd}, Available[idx:]...)...)
	}
}

// VFS returns a command that runs a FUSE file system.
func VFS() *cli.Command {
	return &cli.Command{
		Name:      "vfs",
		Usage:     "Runs a FUSE file system for interacting with datasets",
		UsageText: "aetherfs vfs",
		Action: func(ctx *cli.Context) error {
			return nil
		},
		Hidden:          true,
		HideHelpCommand: true,
	}
}
