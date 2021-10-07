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
