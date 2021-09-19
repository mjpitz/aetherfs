// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package daemons

import (
	"context"
	"io/fs"
	"net/http"
	"strings"

	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
)

type datasetFile struct {
	ctx      context.Context
	dataset  *datasetv1.Dataset
	filePath string
	file     *datasetv1.File
}

func (f *datasetFile) Close() error {
	return nil
}

func (f *datasetFile) Read(p []byte) (n int, err error) {
	panic("implement me")
}

func (f *datasetFile) Seek(offset int64, whence int) (int64, error) {
	panic("implement me")
}

func (f *datasetFile) Readdir(count int) ([]fs.FileInfo, error) {
	seen := make(map[string]bool)
	var infos []fs.FileInfo

	for _, file := range f.dataset.GetFiles() {
		if strings.HasPrefix(file.Name, f.filePath) {
			remaining := strings.TrimPrefix(file.Name, f.filePath)
			remaining = strings.TrimPrefix(remaining, "/")

			idx := strings.Index(remaining, "/")

			switch {
			case idx == -1:
				infos = append(infos, &fileInfo{
					name: remaining,
					file: file,
				})
			case !seen[remaining[:idx]]:
				infos = append(infos, &fileInfo{
					name: remaining[:idx],
				})
				seen[remaining[:idx]] = true
			}
		}
	}

	return infos, nil
}

func (f *datasetFile) Stat() (fs.FileInfo, error) {
	name := f.filePath[strings.LastIndex(f.filePath, "/")+1:]

	return &fileInfo{
		name: name,
		file: f.file,
	}, nil
}

var _ http.File = &datasetFile{}
