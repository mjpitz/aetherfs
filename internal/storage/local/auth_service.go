// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package local

import (
	"context"
	"encoding/json"

	"github.com/zalando/go-keyring"

	"github.com/mjpitz/aetherfs/internal/components"
)

const keychainService = "AetherFS"

type Credentials struct {
	// common name
	Server string `json:"server"`

	components.GRPCClientConfig
}

type Authentications struct{}

func (a *Authentications) Get(ctx context.Context, server string) (creds *Credentials, err error) {
	secret, err := keyring.Get(keychainService, server)
	if err != nil {
		return nil, err
	}

	creds = &Credentials{}
	err = json.Unmarshal([]byte(secret), creds)
	if err != nil {
		return nil, err
	}

	return creds, nil
}

func (a *Authentications) Put(ctx context.Context, creds *Credentials) error {
	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	return keyring.Set(keychainService, creds.Server, string(data))
}

func (a *Authentications) Delete(ctx context.Context, server string) error {
	return keyring.Delete(keychainService, server)
}
