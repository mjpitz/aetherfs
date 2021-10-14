// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package flagset_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/mjpitz/aetherfs/internal/flagset"
)

type Options struct {
	Endpoint    string        `json:"endpoint"    aliases:"e" usage:"the endpoint of the server we're speaking to" default:"default-endpoint"`
	EnableSSL   bool          `json:"enable_ssl"  aliases:"s" usage:"enable encryption between processes"`
	Temperature int           `json:"temperature" aliases:"t"`
}

type Nested struct {
	Options  *Options `json:"options"`
	Repeated []string `json:"repeated"`
}

func TestExtract(t *testing.T) {
	opts := &Options{}
	flags := flagset.Extract(opts)

	require.Len(t, flags, 3)

	{
		flag := flags[0].(*cli.StringFlag)
		require.Equal(t, "endpoint", flag.Name)
		require.Equal(t, "e", flag.Aliases[0])
		require.Equal(t, "ENDPOINT", flag.EnvVars[0])
		require.Equal(t, "the endpoint of the server we're speaking to", flag.Usage)
		require.Equal(t, "default-endpoint", flag.Value)
	}

	{
		flag := flags[1].(*cli.BoolFlag)
		require.Equal(t, "enable_ssl", flag.Name)
		require.Equal(t, "s", flag.Aliases[0])
		require.Equal(t, "ENABLE_SSL", flag.EnvVars[0])
		require.Equal(t, "enable encryption between processes", flag.Usage)
		require.Equal(t, false, flag.Value)
	}

	{
		flag := flags[2].(*cli.IntFlag)
		require.Equal(t, "temperature", flag.Name)
		require.Equal(t, "t", flag.Aliases[0])
		require.Equal(t, "TEMPERATURE", flag.EnvVars[0])
		require.Equal(t, "", flag.Usage)
		require.Equal(t, 0, flag.Value)
	}
}

func TestExtract_Nested(t *testing.T) {
	nested := &Nested{
		Options: &Options{},
	}

	flags := flagset.Extract(nested)

	require.Len(t, flags, 3)

	{
		flag := flags[0].(*cli.StringFlag)
		require.Equal(t, "options_endpoint", flag.Name)
		require.Equal(t, "OPTIONS_ENDPOINT", flag.EnvVars[0])
	}

	{
		flag := flags[1].(*cli.BoolFlag)
		require.Equal(t, "options_enable_ssl", flag.Name)
		require.Equal(t, "OPTIONS_ENABLE_SSL", flag.EnvVars[0])
	}

	{
		flag := flags[2].(*cli.IntFlag)
		require.Equal(t, "options_temperature", flag.Name)
		require.Equal(t, "OPTIONS_TEMPERATURE", flag.EnvVars[0])
	}
}
