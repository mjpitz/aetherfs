// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package storage

import (
	"fmt"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	"github.com/mjpitz/aetherfs/internal/storage/s3"
)

type Config struct {
	Driver string `json:"driver" usage:"configure how information is stored"`

	S3 s3.Config `json:"s3"`
}

type Stores struct {
	BlockAPIServer   blockv1.BlockAPIServer
	DatasetAPIServer datasetv1.DatasetAPIServer
}

func ObtainStores(cfg Config) (*Stores, error) {
	var blockAPI blockv1.BlockAPIServer
	var datasetAPI datasetv1.DatasetAPIServer
	var err error

	switch cfg.Driver {
	case "s3":
		blockAPI, datasetAPI, err = s3.ObtainStores(cfg.S3)

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
