// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package daemons

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
)

type datasetService struct {
	datasetv1.UnsafeDatasetAPIServer
}

func (d *datasetService) List(ctx context.Context, request *datasetv1.ListRequest) (*datasetv1.ListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (d *datasetService) ListTags(ctx context.Context, request *datasetv1.ListTagsRequest) (*datasetv1.ListTagsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (d *datasetService) Lookup(ctx context.Context, request *datasetv1.LookupRequest) (*datasetv1.LookupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (d *datasetService) Publish(ctx context.Context, request *datasetv1.PublishRequest) (*datasetv1.PublishResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (d *datasetService) Subscribe(server datasetv1.DatasetAPI_SubscribeServer) error {
	return status.Errorf(codes.Unimplemented, "unimplemented")
}

var _ datasetv1.DatasetAPIServer = &datasetService{}

