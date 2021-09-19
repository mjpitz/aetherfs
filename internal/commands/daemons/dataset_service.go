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

var (
	mockDatasets = map[string]map[string]*datasetv1.Dataset{
		"@scoped/dataset": {
			"latest": {
				Files: []*datasetv1.File{
					{
						Name: "directory/file-1.csv",
					},
					{
						Name: "directory/file-2.json",
					},
					{
						Name: "README.md",
					},
				},
				BlockSize: 0,
				Blocks:    []string{},
			},
			"stable": {
				Files: []*datasetv1.File{
					{
						Name: "directory/file-1.csv",
					},
					{
						Name: "directory/file-2.json",
					},
					{
						Name: "README.md",
					},
				},
				BlockSize: 0,
				Blocks:    []string{},
			},
		},
		"dataset": {
			"latest": {
				Files: []*datasetv1.File{
					{
						Name: "file-1.csv",
					},
					{
						Name: "file-2.json",
					},
					{
						Name: "README.md",
					},
				},
				BlockSize: 0,
				Blocks:    []string{},
			},
			"stable": {
				Files: []*datasetv1.File{
					{
						Name: "file-1.csv",
					},
					{
						Name: "file-2.json",
					},
					{
						Name: "README.md",
					},
				},
				BlockSize: 0,
				Blocks:    []string{},
			},
			"next": {
				Files: []*datasetv1.File{
					{
						Name: "file-1.csv",
					},
					{
						Name: "file-2.json",
					},
					{
						Name: "README.md",
					},
				},
				BlockSize: 0,
				Blocks:    []string{},
			},
		},
	}
)

type datasetService struct {
	datasetv1.UnsafeDatasetAPIServer
}

func (d *datasetService) List(ctx context.Context, request *datasetv1.ListRequest) (*datasetv1.ListResponse, error) {
	resp := &datasetv1.ListResponse{}
	for dataset := range mockDatasets {
		resp.Datasets = append(resp.Datasets, dataset)
	}

	return resp, nil
}

func (d *datasetService) ListTags(ctx context.Context, request *datasetv1.ListTagsRequest) (*datasetv1.ListTagsResponse, error) {
	tags, ok := mockDatasets[request.GetName()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "not found")
	}

	resp := &datasetv1.ListTagsResponse{}
	for tag := range tags {
		resp.Tags = append(resp.Tags, &datasetv1.Tag{
			Name:    request.GetName(),
			Version: tag,
		})
	}

	return resp, nil
}

func (d *datasetService) Lookup(ctx context.Context, request *datasetv1.LookupRequest) (*datasetv1.LookupResponse, error) {
	tags, ok := mockDatasets[request.GetTag().GetName()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "not found")
	}

	resp := &datasetv1.LookupResponse{}
	resp.Dataset, ok = tags[request.GetTag().GetVersion()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "not found")
	}

	return resp, nil
}

func (d *datasetService) Publish(ctx context.Context, request *datasetv1.PublishRequest) (*datasetv1.PublishResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (d *datasetService) Subscribe(server datasetv1.DatasetAPI_SubscribeServer) error {
	return status.Errorf(codes.Unimplemented, "unimplemented")
}

var _ datasetv1.DatasetAPIServer = &datasetService{}
