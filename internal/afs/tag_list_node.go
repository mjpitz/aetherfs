// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package afs

import (
	"os"
	"strings"

	"github.com/spf13/afero"

	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
)

type tagListNode struct {
	filePath string
	tagList  []*datasetv1.Tag
}

func (d *tagListNode) Stat() (os.FileInfo, error) {
	name := d.filePath[strings.LastIndex(d.filePath, "/")+1:]

	return &fileInfo{
		name: name,
	}, nil
}

func (d *tagListNode) Readdir(count int) ([]os.FileInfo, error) {
	var infos []os.FileInfo
	for _, tag := range d.tagList {
		infos = append(infos, &fileInfo{
			name: tag.Version,
		})
	}
	return infos, nil
}

func (d *tagListNode) Readdirnames(count int) ([]string, error) {
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

func (d *tagListNode) Name() string {
	info, _ := d.Stat()
	return info.Name()
}

// invalid for directory

func (d *tagListNode) Read(p []byte) (n int, err error) {
	return 0, os.ErrInvalid
}

func (d *tagListNode) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, os.ErrInvalid
}

func (d *tagListNode) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (d *tagListNode) Sync() error {
	return os.ErrInvalid
}

func (d *tagListNode) Close() error {
	return os.ErrInvalid
}

func (d *tagListNode) Write(p []byte) (n int, err error) {
	return 0, os.ErrInvalid
}

func (d *tagListNode) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, os.ErrInvalid
}

func (d *tagListNode) Truncate(size int64) error {
	return os.ErrInvalid
}

func (d *tagListNode) WriteString(s string) (ret int, err error) {
	return 0, os.ErrInvalid
}

var _ afero.File = &tagListNode{}
