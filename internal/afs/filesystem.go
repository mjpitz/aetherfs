// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package afs

import (
	"context"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/afero"
	"go.uber.org/zap"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
)

type FileSystem struct {
	Context context.Context

	BlockAPI   blockv1.BlockAPIClient
	DatasetAPI datasetv1.DatasetAPIClient
}

func (f *FileSystem) Name() string {
	return "AetherFS"
}

func (f *FileSystem) Stat(name string) (os.FileInfo, error) {
	file, err := f.Open(name)
	if err != nil {
		return nil, err
	}

	return file.Stat()
}

// openDatasetList renders top level nodes that list datasets within the File system.
func (f *FileSystem) openDatasetList(scope string) (afero.File, error) {
	listResp, err := f.DatasetAPI.List(f.Context, &datasetv1.ListRequest{})
	if err != nil {
		return nil, translateError(err)
	}

	datasets := make([]string, 0)
	for _, dataset := range listResp.GetDatasets() {
		if scope == "" || strings.HasPrefix(dataset.GetName(), scope+"/") {
			datasets = append(datasets, dataset.GetName())
		}
	}

	var filePath string
	if scope != "" {
		if len(datasets) == 0 {
			return nil, os.ErrNotExist
		}
		filePath = scope + "/"
	}

	return &datasetListNode{
		filePath:    filePath,
		datasetList: datasets,
	}, nil
}

// openTagList renders the list of tags for the provided dataset.
func (f *FileSystem) openTagList(scope, dataset string) (afero.File, error) {
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

// openDatasetFile renders a files within a given tagged dataset.
func (f *FileSystem) openDatasetFile(scope, dataset, tag, filePath string) (afero.File, error) {
	if scope != "" {
		dataset = scope + "/" + dataset
	}

	// load dataset
	// CurrentPath may be a directory (prefix) or DatasetFile within the given dataset
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
		return &DatasetFile{
			Context:     f.Context,
			BlockAPI:    f.BlockAPI,
			Dataset:     resp.GetDataset(),
			CurrentPath: filePath,
			File:        requestedFile, // maybe nil
		}, nil
	}

	return nil, os.ErrNotExist
}

func (f *FileSystem) Open(name string) (afero.File, error) {
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
		return f.openDatasetList(scope)
	} else if tag == "" {
		return f.openTagList(scope, dataset)
	}

	return f.openDatasetFile(scope, dataset, tag, filePath)
}

func (f *FileSystem) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	mask := os.O_WRONLY | os.O_RDWR | os.O_APPEND | os.O_CREATE | os.O_TRUNC
	if flag&mask != 0 {
		return nil, syscall.EPERM
	}

	return f.Open(name)
}

// unsupported, not read only

func (f *FileSystem) Mkdir(name string, perm os.FileMode) error {
	return syscall.EPERM
}

func (f *FileSystem) MkdirAll(path string, perm os.FileMode) error {
	return syscall.EPERM
}

func (f *FileSystem) Create(name string) (afero.File, error) {
	return nil, syscall.EPERM
}

func (f *FileSystem) Remove(name string) error {
	return syscall.EPERM
}

func (f *FileSystem) RemoveAll(path string) error {
	return syscall.EPERM
}

func (f *FileSystem) Rename(oldname, newname string) error {
	return syscall.EPERM
}

func (f *FileSystem) Chmod(name string, mode os.FileMode) error {
	return syscall.EPERM
}

func (f *FileSystem) Chown(name string, uid, gid int) error {
	return syscall.EPERM
}

func (f *FileSystem) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return syscall.EPERM
}
