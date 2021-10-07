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

package flagset_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/mjpitz/aetherfs/internal/flagset"
)

type Options struct {
	Endpoint    string        `json:"endpoint"    aliases:"e" usage:"the endpoint of the server we're speaking to"`
	EnableSSL   bool          `json:"enable_ssl"  aliases:"s" usage:"enable encryption between processes"`
	ValidFor    time.Duration `json:"valid_for"   aliases:"v" usage:"how long tokens are good for before expiring"`
	Temperature int           `json:"temperature" aliases:"t"`
	BlockSize   uint          `json:"block_size"`
}

type Nested struct {
	Options  *Options `json:"options"`
	Repeated []string `json:"repeated"`
}

func TestExtract(t *testing.T) {
	opts := &Options{
		Endpoint: "default-endpoint",
		ValidFor: time.Minute,
	}
	flags := flagset.Extract(opts)

	require.Len(t, flags, 5)

	{
		flag := flags[0].(*cli.GenericFlag)
		require.Equal(t, "endpoint", flag.Name)
		require.Equal(t, "e", flag.Aliases[0])
		require.Equal(t, "ENDPOINT", flag.EnvVars[0])
		require.Equal(t, "the endpoint of the server we're speaking to", flag.Usage)
		require.Equal(t, "default-endpoint", flag.GetValue())
	}

	{
		flag := flags[1].(*cli.GenericFlag)
		require.Equal(t, "enable_ssl", flag.Name)
		require.Equal(t, "s", flag.Aliases[0])
		require.Equal(t, "ENABLE_SSL", flag.EnvVars[0])
		require.Equal(t, "enable encryption between processes", flag.Usage)
		require.Equal(t, "false", flag.GetValue())
	}

	{
		flag := flags[2].(*cli.GenericFlag)
		require.Equal(t, "valid_for", flag.Name)
		require.Equal(t, "v", flag.Aliases[0])
		require.Equal(t, "VALID_FOR", flag.EnvVars[0])
		require.Equal(t, "how long tokens are good for before expiring", flag.Usage)
		require.Equal(t, "1m0s", flag.GetValue())
	}

	{
		flag := flags[3].(*cli.GenericFlag)
		require.Equal(t, "temperature", flag.Name)
		require.Equal(t, "t", flag.Aliases[0])
		require.Equal(t, "TEMPERATURE", flag.EnvVars[0])
		require.Equal(t, "", flag.Usage)
		require.Equal(t, "0", flag.GetValue())
	}

	{
		flag := flags[4].(*cli.GenericFlag)
		require.Equal(t, "block_size", flag.Name)
		require.Equal(t, "BLOCK_SIZE", flag.EnvVars[0])
		require.Equal(t, "", flag.Usage)
		require.Equal(t, "0", flag.GetValue())
	}
}

func TestExtract_Nested(t *testing.T) {
	nested := &Nested{
		Options: &Options{
			Endpoint: "default-endpoint",
			ValidFor: time.Minute,
		},
	}

	flags := flagset.Extract(nested)

	require.Len(t, flags, 6)

	{
		flag := flags[0].(*cli.GenericFlag)
		require.Equal(t, "options_endpoint", flag.Name)
		require.Equal(t, "OPTIONS_ENDPOINT", flag.EnvVars[0])
	}

	{
		flag := flags[1].(*cli.GenericFlag)
		require.Equal(t, "options_enable_ssl", flag.Name)
		require.Equal(t, "OPTIONS_ENABLE_SSL", flag.EnvVars[0])
	}

	{
		flag := flags[2].(*cli.GenericFlag)
		require.Equal(t, "options_valid_for", flag.Name)
		require.Equal(t, "OPTIONS_VALID_FOR", flag.EnvVars[0])
	}

	{
		flag := flags[3].(*cli.GenericFlag)
		require.Equal(t, "options_temperature", flag.Name)
		require.Equal(t, "OPTIONS_TEMPERATURE", flag.EnvVars[0])
	}

	{
		flag := flags[4].(*cli.GenericFlag)
		require.Equal(t, "options_block_size", flag.Name)
		require.Equal(t, "OPTIONS_BLOCK_SIZE", flag.EnvVars[0])
	}

	{
		flag := flags[5].(*cli.GenericFlag)
		require.Equal(t, "repeated", flag.Name)
		require.Equal(t, "REPEATED", flag.EnvVars[0])
	}
}
