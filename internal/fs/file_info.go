// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package fs

import (
	"io/fs"
	"time"

	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
)

type fileInfo struct {
	name string
	file *datasetv1.File
}

func (f *fileInfo) Name() string {
	return f.name
}

func (f *fileInfo) Size() int64 {
	if f.file == nil {
		return 0
	}

	return f.file.GetSize()
}

func (f *fileInfo) Mode() fs.FileMode {
	if f.file == nil {
		return 0444 & fs.ModeDir
	}

	return 0444
}

func (f *fileInfo) ModTime() time.Time {
	if f.file == nil {
		return time.Now()
	}

	return f.file.GetLastModified().AsTime()
}

func (f *fileInfo) IsDir() bool {
	return f.file == nil
}

func (f *fileInfo) Sys() interface{} {
	return nil
}

var _ fs.FileInfo = &fileInfo{}
