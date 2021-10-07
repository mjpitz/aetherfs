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

package s3

import (
	"bytes"
	"context"
	"encoding/json"
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
		if info.Err != nil {
			return nil, status.Errorf(codes.Internal, "failed to list datasets")
		}

		key := strings.TrimPrefix(info.Key, "datasets/")
		key = strings.TrimSuffix(key, "/")

		if strings.HasPrefix(key, "@") {
			scopes = append(scopes, key)
		} else {
			resp.Datasets = append(resp.Datasets, key)
		}
	}

	for _, scope := range scopes {
		opts := minio.ListObjectsOptions{
			Prefix: "datasets/" + scope + "/",
		}

		for info := range d.s3Client.ListObjects(ctx, d.bucketName, opts) {
			if info.Err != nil {
				return nil, status.Errorf(codes.Internal, "failed to list datasets")
			}

			key := strings.TrimPrefix(info.Key, "datasets/")
			key = strings.TrimSuffix(key, "/")

			resp.Datasets = append(resp.Datasets, key)
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
		if info.Err != nil {
			return nil, status.Errorf(codes.Internal, "failed to list tags")
		}

		key := strings.TrimPrefix(info.Key, opts.Prefix)
		key = strings.TrimSuffix(key, "/")

		resp.Tags = append(resp.Tags, &datasetv1.Tag{
			Name:    request.Name,
			Version: key,
		})
	}

	return resp, nil
}

func (d *datasetService) Lookup(ctx context.Context, request *datasetv1.LookupRequest) (*datasetv1.LookupResponse, error) {
	objectKey := "datasets/" + request.Tag.Name + "/" + request.Tag.Version

	obj, err := d.s3Client.GetObject(ctx, d.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to lookup dataset")
	}

	data, err := ioutil.ReadAll(obj)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read dataset")
	}

	dataset := &datasetv1.Dataset{}

	err = json.Unmarshal(data, dataset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmarshal dataset")
	}

	return &datasetv1.LookupResponse{
		Dataset: dataset,
	}, nil
}

func (d *datasetService) Publish(ctx context.Context, request *datasetv1.PublishRequest) (*datasetv1.PublishResponse, error) {
	data, err := json.Marshal(request.Dataset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal dataset")
	}

	for _, tag := range request.Tags {
		objectKey := "datasets/" + tag.Name + "/" + tag.Version

		_, err = d.s3Client.PutObject(ctx, d.bucketName, objectKey,
			bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})

		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to write tag")
		}
	}

	return &datasetv1.PublishResponse{}, nil
}

func (d *datasetService) Subscribe(call datasetv1.DatasetAPI_SubscribeServer) error {
	return status.Errorf(codes.Unimplemented, "unimplemented")
}

var _ datasetv1.DatasetAPIServer = &datasetService{}
