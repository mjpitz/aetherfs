// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package daemons

import (
	"io/fs"
	"net/http"
	"os"
	"strings"
)

type datasetListNode struct {
	filePath    string
	datasetList []string
}

func (d *datasetListNode) Close() error {
	return os.ErrInvalid
}

func (d *datasetListNode) Read(p []byte) (n int, err error) {
	return 0, os.ErrInvalid
}

func (d *datasetListNode) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (d *datasetListNode) Readdir(count int) ([]fs.FileInfo, error) {
	var infos []fs.FileInfo

	for _, dataset := range d.datasetList {
		parts := strings.Split(dataset, "/")

		if d.filePath == "" {
			infos = append(infos, &fileInfo{
				name: parts[0],
			})
		} else if strings.HasPrefix(dataset, d.filePath) {
			// we don't technically need this, but its a good safeguard
			infos = append(infos, &fileInfo{
				name: parts[1],
			})
		}
	}

	return infos, nil
}

func (d *datasetListNode) Stat() (fs.FileInfo, error) {
	name := d.filePath[strings.LastIndex(d.filePath, "/")+1:]

	return &fileInfo{
		name: name,
	}, nil
}

var _ http.File = &datasetListNode{}
