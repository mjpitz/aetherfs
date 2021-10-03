// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"

	"github.com/minio/minio-go/v7"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
)

type datasetService struct {
	datasetv1.UnsafeDatasetAPIServer

	s3Client   *minio.Client
	bucketName string
}

func (d *datasetService) List(ctx context.Context, request *datasetv1.ListRequest) (*datasetv1.ListResponse, error) {
	scopes := make([]string, 0)
	resp := &datasetv1.ListResponse{}

	opts := minio.ListObjectsOptions{
		Prefix: "datasets/",
	}

	for info := range d.s3Client.ListObjects(ctx, d.bucketName, opts) {
		switch {
		case info.Err == io.EOF:
			break
		case info.Err != nil:
			return nil, status.Errorf(codes.Internal, "internal server error")
		}

		if strings.HasPrefix(info.Key, "@") {
			scopes = append(scopes, info.Key)
		} else {
			resp.Datasets = append(resp.Datasets, info.Key)
		}
	}

	for _, scope := range scopes {
		opts := minio.ListObjectsOptions{
			Prefix: "datasets/" + scope + "/",
		}

		for info := range d.s3Client.ListObjects(ctx, d.bucketName, opts) {
			switch {
			case info.Err == io.EOF:
				break
			case info.Err != nil:
				return nil, status.Errorf(codes.Internal, "internal server error")
			}

			resp.Datasets = append(resp.Datasets, scope+"/"+info.Key)
		}
	}

	return resp, nil
}

func (d *datasetService) ListTags(ctx context.Context, request *datasetv1.ListTagsRequest) (*datasetv1.ListTagsResponse, error) {
	opts := minio.ListObjectsOptions{
		Prefix: "datasets/" + request.Name + "/",
	}

	resp := &datasetv1.ListTagsResponse{}
	for info := range d.s3Client.ListObjects(ctx, d.bucketName, opts) {
		switch {
		case info.Err == io.EOF:
			break
		case info.Err != nil:
			return nil, status.Errorf(codes.Internal, "internal server error")
		}

		resp.Tags = append(resp.Tags, &datasetv1.Tag{
			Name:    request.Name,
			Version: info.Key,
		})
	}

	return resp, nil
}

func (d *datasetService) Lookup(ctx context.Context, request *datasetv1.LookupRequest) (*datasetv1.LookupResponse, error) {
	objectKey := "datasets/" + request.Tag.Name + "/" + request.Tag.Version

	obj, err := d.s3Client.GetObject(ctx, d.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	data, err := ioutil.ReadAll(obj)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	dataset := &datasetv1.Dataset{}

	err = json.Unmarshal(data, dataset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	return &datasetv1.LookupResponse{
		Dataset: dataset,
	}, nil
}

func (d *datasetService) Publish(ctx context.Context, request *datasetv1.PublishRequest) (*datasetv1.PublishResponse, error) {
	data, err := json.Marshal(request.Dataset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	for _, tag := range request.Tags {
		objectKey := "datasets/" + tag.Name + "/" + tag.Version

		_, err = d.s3Client.PutObject(ctx, d.bucketName, objectKey,
			bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})

		if err != nil {
			return nil, status.Errorf(codes.Internal, "")
		}
	}

	return &datasetv1.PublishResponse{}, nil
}

func (d *datasetService) Subscribe(call datasetv1.DatasetAPI_SubscribeServer) error {
	return status.Errorf(codes.Unimplemented, "unimplemented")
}

var _ datasetv1.DatasetAPIServer = &datasetService{}
