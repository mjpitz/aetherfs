// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package storage

import (
	"context"
	"fmt"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	"github.com/mjpitz/aetherfs/internal/storage/proxy"
	"github.com/mjpitz/aetherfs/internal/storage/s3"
)

type Config struct {
	Driver string `json:"driver" usage:"configure how information is stored" default:"s3"`

	S3    s3.Config    `json:"s3"`
	Proxy proxy.Config `json:"proxy"`
}

type Stores struct {
	BlockAPIServer   blockv1.BlockAPIServer
	DatasetAPIServer datasetv1.DatasetAPIServer
}

func ObtainStores(ctx context.Context, cfg Config) (*Stores, error) {
	var blockAPI blockv1.BlockAPIServer
	var datasetAPI datasetv1.DatasetAPIServer
	var err error

	switch cfg.Driver {
	case "s3":
		blockAPI, datasetAPI, err = s3.ObtainStores(ctx, cfg.S3)
	case "proxy":
		blockAPI, datasetAPI, err = proxy.ObtainStores(ctx, cfg.Proxy)
	case "", "none":
		return nil, nil
	default:
		err = fmt.Errorf("invalid driver: %s", cfg.Driver)
	}

	if err != nil {
		return nil, err
	}

	return &Stores{
		BlockAPIServer:   blockAPI,
		DatasetAPIServer: datasetAPI,
	}, nil
}
