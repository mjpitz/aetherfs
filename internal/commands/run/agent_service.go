// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package run

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	agentv1 "github.com/mjpitz/aetherfs/api/aetherfs/agent/v1"
	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
)

type agentService struct {
	agentv1.UnsafeAgentAPIServer

	blockAPI   blockv1.BlockAPIClient
	datasetAPI datasetv1.DatasetAPIClient
}

func (a *agentService) Publish(ctx context.Context, request *agentv1.PublishRequest) (*agentv1.PublishResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (a *agentService) Subscribe(server agentv1.AgentAPI_SubscribeServer) error {
	return status.Errorf(codes.Unimplemented, "unimplemented")
}

func (a *agentService) GracefulShutdown(ctx context.Context, request *agentv1.GracefulShutdownRequest) (*agentv1.GracefulShutdownResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

var _ agentv1.AgentAPIServer = &agentService{}
