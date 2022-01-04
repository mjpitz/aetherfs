// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package proxy

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
)

type datasetService struct {
	datasetv1.UnsafeDatasetAPIServer

	delegate datasetv1.DatasetAPIClient
}

func (d *datasetService) List(ctx context.Context, request *datasetv1.ListRequest) (*datasetv1.ListResponse, error) {
	return d.delegate.List(ctx, request)
}

func (d *datasetService) ListTags(ctx context.Context, request *datasetv1.ListTagsRequest) (*datasetv1.ListTagsResponse, error) {
	return d.delegate.ListTags(ctx, request)
}

func (d *datasetService) Lookup(ctx context.Context, request *datasetv1.LookupRequest) (*datasetv1.LookupResponse, error) {
	return d.delegate.Lookup(ctx, request)
}

func (d *datasetService) Publish(ctx context.Context, request *datasetv1.PublishRequest) (*datasetv1.PublishResponse, error) {
	return d.delegate.Publish(ctx, request)
}

func (d *datasetService) Subscribe(call datasetv1.DatasetAPI_SubscribeServer) error {
	return status.Errorf(codes.Unimplemented, "unimplemented")
}

var _ datasetv1.DatasetAPIServer = &datasetService{}
