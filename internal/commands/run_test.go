// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package commands_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mjpitz/aetherfs/internal/commands"
)

const configSnapshot = `{
  "config_file": "",
  "port": 0,
  "tls": {
    "enable": false,
    "cert_path": "",
    "ca_file": "",
    "cert_file": "",
    "key_file": "",
    "reload_interval": 0
  },
  "auth_type": "",
  "oidc": {
    "issuer": {
      "server_url": "",
      "certificate_authority": ""
    }
  },
  "storage": {
    "driver": "",
    "s3": {
      "endpoint": "",
      "tls": {
        "enable": false,
        "cert_path": "",
        "ca_file": "",
        "cert_file": "",
        "key_file": "",
        "reload_interval": 0
      },
      "access_key_id": "",
      "secret_access_key": "",
      "region": "",
      "bucket": ""
    }
  }
}`

func TestConfigSnapshot(t *testing.T) {
	data, err := json.MarshalIndent(commands.RunConfig{}, "", "  ")
	require.NoError(t, err)
	require.Equal(t, configSnapshot, string(data))
}
