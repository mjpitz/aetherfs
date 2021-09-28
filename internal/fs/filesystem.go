// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package fs

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
)

type FileSystem struct {
	Context context.Context

	BlockAPI   blockv1.BlockAPIClient
	DatasetAPI datasetv1.DatasetAPIClient
}

// renderDatasetList renders top level nodes that list datasets within the file system.
func (f *FileSystem) renderDatasetList(scope string) (http.File, error) {
	listResp, err := f.DatasetAPI.List(f.Context, &datasetv1.ListRequest{})
	if err != nil {
		return nil, translateError(err)
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

	return &datasetListNode{
		filePath:    filePath,
		datasetList: datasets,
	}, nil
}

// renderTagList renders the list of tags for the provided dataset.
func (f *FileSystem) renderTagList(scope, dataset string) (http.File, error) {
	if scope != "" {
		dataset = scope + "/" + dataset
	}

	listTagsResp, err := f.DatasetAPI.ListTags(f.Context, &datasetv1.ListTagsRequest{
		Name: dataset,
	})

	if err != nil {
		return nil, translateError(err)
	}

	return &tagListNode{
		filePath: dataset,
		tagList:  listTagsResp.GetTags(),
	}, nil
}

// renderDatasetFile renders a files within a given tagged dataset.
func (f *FileSystem) renderDatasetFile(scope, dataset, tag, filePath string) (http.File, error) {
	if scope != "" {
		dataset = scope + "/" + dataset
	}

	// load dataset
	// filePath may be a directory (prefix) or datasetFile within the given dataset
	resp, err := f.DatasetAPI.Lookup(f.Context, &datasetv1.LookupRequest{
		Tag: &datasetv1.Tag{
			Name:    dataset,
			Version: tag,
		},
	})

	if err != nil {
		return nil, translateError(err)
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
			ctx:      f.Context,
			blockAPI: f.BlockAPI,
			dataset:  resp.GetDataset(),
			filePath: filePath,
			file:     requestedFile,		// maybe nil
		}, nil
	}

	return nil, os.ErrNotExist
}

// Open is called with the full datasetFile path
// 1.631977188164028e+09   info    daemons/filesystem.go:18        open    {"name": "/test"}
// 1.631977204080589e+09   info    daemons/filesystem.go:18        open    {"name": "/test/path"}
// 1.6319772103438861e+09  info    daemons/filesystem.go:18        open    {"name": "/test/path.jpg"}
func (f *FileSystem) Open(name string) (http.File, error) {
	var parts []string

	// in some cases, go will append index.html to a file
	// this is a gross hack to prevent that from happening
	name = strings.TrimSuffix(name, "/index.html")

	ctxzap.Extract(f.Context).Info("open", zap.String("path", name))

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

	ctxzap.Extract(f.Context).Info("route", zap.Strings("parts", parts))

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

var _ http.FileSystem = &FileSystem{}