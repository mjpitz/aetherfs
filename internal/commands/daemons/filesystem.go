// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package daemons

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
)

type fileSystem struct {
	ctx context.Context

	blockAPI   blockv1.BlockAPIClient
	datasetAPI datasetv1.DatasetAPIClient
}

// translateError takes in an arbitrary error and attempts to convert it to a more meaningful error code.
func (f *fileSystem) translateError(err error) (http.File, error) {
	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case codes.Unauthenticated:
			return nil, os.ErrPermission
		case codes.NotFound:
			return nil, os.ErrNotExist
		case codes.DeadlineExceeded:
			return nil, os.ErrDeadlineExceeded
		}
	}

	return nil, err
}

// renderDatasetList renders top level nodes that list datasets within the file system.
func (f *fileSystem) renderDatasetList(scope string) (http.File, error) {
	listResp, err := f.datasetAPI.List(f.ctx, &datasetv1.ListRequest{})
	if err != nil {
		return f.translateError(err)
	}

	datasets := listResp.GetDatasets()

	filePath := ""
	if scope != "" {
		filePath = scope + "/"

		filteredDatasets := make([]string, 0, len(datasets))
		for _, dataset := range datasets {
			if strings.HasPrefix(dataset, filePath) {
				filteredDatasets = append(filteredDatasets, dataset)
			}
		}

		if len(filteredDatasets) == 0 {
			return nil, os.ErrNotExist
		}

		datasets = filteredDatasets
	}

	ctxzap.Extract(f.ctx).Info("dataset list", zap.String("filePath", filePath), zap.Strings("datasets", datasets))

	return &datasetListNode{
		filePath:    filePath,
		datasetList: datasets,
	}, nil
}

// renderTagList renders the list of tags for the provided dataset.
func (f *fileSystem) renderTagList(scope, dataset string) (http.File, error) {
	if scope != "" {
		dataset = scope + "/" + dataset
	}

	listTagsResp, err := f.datasetAPI.ListTags(f.ctx, &datasetv1.ListTagsRequest{
		Name: dataset,
	})

	if err != nil {
		return f.translateError(err)
	}

	return &tagListNode{
		filePath: dataset,
		tagList:  listTagsResp.GetTags(),
	}, nil
}

// renderDatasetFile renders a files within a given tagged dataset.
func (f *fileSystem) renderDatasetFile(scope, dataset, tag, filePath string) (http.File, error) {
	if scope != "" {
		dataset = scope + "/" + dataset
	}

	// load dataset
	// filePath may be a directory (prefix) or datasetFile within the given dataset
	resp, err := f.datasetAPI.Lookup(f.ctx, &datasetv1.LookupRequest{
		Tag: &datasetv1.Tag{
			Name:    dataset,
			Version: tag,
		},
	})

	if err != nil {
		return f.translateError(err)
	}

	var requestedFile *datasetv1.File
	isDirectory := false

	for _, file := range resp.GetDataset().GetFiles() {
		if file.Name == filePath {
			requestedFile = file
		}

		isDirectory = isDirectory || strings.HasPrefix(file.Name, filePath)
	}

	if requestedFile != nil || isDirectory {
		return &datasetFile{
			ctx:      f.ctx,
			dataset:  resp.GetDataset(),
			filePath: filePath,
			file:     requestedFile,
		}, nil
	}

	return nil, os.ErrNotExist
}

// Open is called with the full datasetFile path
// 1.631977188164028e+09   info    daemons/filesystem.go:18        open    {"name": "/test"}
// 1.631977204080589e+09   info    daemons/filesystem.go:18        open    {"name": "/test/path"}
// 1.6319772103438861e+09  info    daemons/filesystem.go:18        open    {"name": "/test/path.jpg"}
func (f *fileSystem) Open(name string) (http.File, error) {
	var parts []string

	// in some cases, go will append index.html to a file
	// this is a gross hack to prevent that from happening
	name = strings.TrimSuffix(name, "/index.html")

	ctxzap.Extract(f.ctx).Info("open", zap.String("path", name))

	if strings.HasPrefix(strings.TrimPrefix(name, "/"), "@") {
		parts = strings.SplitN(name, "/", 5)
		parts = parts[1:]

	} else {
		// exploits leading / to end up with an empty scope
		parts = strings.SplitN(name, "/", 4)
	}

	// fill in any missing parts
	for len(parts) < 4 {
		parts = append(parts, "")
	}

	// make sure all prior parts are provided
	// otherwise, this is an invalid request
	// for example
	// - you cannot provide a filePath without first specifying a tag or dataset
	// - you cannot provide a tag without first specifying a dataset

	provided := false
	for i := len(parts) - 1; i > 0; i-- {
		if provided && parts[i] == "" {
			return nil, os.ErrInvalid
		}

		provided = provided || parts[i] != ""
	}

	ctxzap.Extract(f.ctx).Info("route", zap.Strings("parts", parts))

	scope := parts[0]
	dataset := parts[1]
	tag := parts[2]
	filePath := parts[3]

	if dataset == "" {
		return f.renderDatasetList(scope)
	} else if tag == "" {
		return f.renderTagList(scope, dataset)
	}

	return f.renderDatasetFile(scope, dataset, tag, filePath)
}

var _ http.FileSystem = &fileSystem{}
