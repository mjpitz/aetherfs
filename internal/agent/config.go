// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package agent

type Config struct {
	Enable bool `json:"enable" usage:"enable the agent API"`

	Shutdown struct {
		Enable bool `json:"enable" usage:"enables the agent API to initiate a shutdown"`
	} `json:"shutdown"`
}
