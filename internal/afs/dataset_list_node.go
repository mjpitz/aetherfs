// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package afs

import (
	"os"
	"strings"

	"github.com/spf13/afero"
)

type datasetListNode struct {
	filePath    string
	datasetList []string
}

func (d *datasetListNode) Stat() (os.FileInfo, error) {
	name := d.filePath[strings.LastIndex(d.filePath, "/")+1:]

	return &fileInfo{
		name: name,
	}, nil
}

func (d *datasetListNode) Readdir(count int) ([]os.FileInfo, error) {
	var infos []os.FileInfo

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

func (d *datasetListNode) Readdirnames(count int) ([]string, error) {
	infos, err := d.Readdir(count)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = info.Name()
	}

	return names, nil
}

func (d *datasetListNode) Name() string {
	info, _ := d.Stat()
	return info.Name()
}

// invalid for directory

func (d *datasetListNode) Read(p []byte) (n int, err error) {
	return 0, os.ErrInvalid
}

func (d *datasetListNode) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, os.ErrInvalid
}

func (d *datasetListNode) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (d *datasetListNode) Sync() error {
	return os.ErrInvalid
}

func (d *datasetListNode) Close() error {
	return os.ErrInvalid
}

func (d *datasetListNode) Write(p []byte) (n int, err error) {
	return 0, os.ErrInvalid
}

func (d *datasetListNode) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, os.ErrInvalid
}

func (d *datasetListNode) Truncate(size int64) error {
	return os.ErrInvalid
}

func (d *datasetListNode) WriteString(s string) (ret int, err error) {
	return 0, os.ErrInvalid
}

var _ afero.File = &datasetListNode{}
