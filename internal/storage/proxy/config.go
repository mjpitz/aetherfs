// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package proxy

import (
	"context"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	"github.com/mjpitz/aetherfs/internal/components"
	"github.com/mjpitz/myago/livetls"
)

type Config struct {
	Target string         `json:"target" usage:"address the grpc client should dial"`
	TLS    livetls.Config `json:"tls"`
}

func ObtainStores(ctx context.Context, cfg Config) (blockv1.BlockAPIServer, datasetv1.DatasetAPIServer, error) {
	conn, err := components.GRPCClient(ctx, components.GRPCClientConfig{
		Target: cfg.Target,
		TLS:    cfg.TLS,
	})
	if err != nil {
		return nil, nil, err
	}

	blockSvc := &blockService{
		delegate: blockv1.NewBlockAPIClient(conn),
	}

	datasetSvc := &datasetService{
		delegate: datasetv1.NewDatasetAPIClient(conn),
	}

	return blockSvc, datasetSvc, nil
}
