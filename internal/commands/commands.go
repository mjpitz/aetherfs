// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package commands

import (
	"github.com/urfave/cli/v2"
)

// Available contains an array of commands that are available. Some commands are not available on some
// machine architectures.
var Available = []*cli.Command{
	Auth(),
	Pull(),
	Push(),
	Run(),
	Version(),
}
