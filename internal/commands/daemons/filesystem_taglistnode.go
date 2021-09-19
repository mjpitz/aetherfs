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

	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
)

type tagListNode struct {
	filePath string
	tagList  []*datasetv1.Tag
}

func (t *tagListNode) Close() error {
	return os.ErrInvalid
}

func (t *tagListNode) Read(p []byte) (n int, err error) {
	return 0, os.ErrInvalid
}

func (t *tagListNode) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (t *tagListNode) Readdir(count int) ([]fs.FileInfo, error) {
	var infos []fs.FileInfo
	for _, tag := range t.tagList {
		infos = append(infos, &fileInfo{
			name: tag.Version,
		})
	}
	return infos, nil
}

func (t *tagListNode) Stat() (fs.FileInfo, error) {
	name := t.filePath[strings.LastIndex(t.filePath, "/")+1:]

	return &fileInfo{
		name: name,
	}, nil
}

var _ http.File = &tagListNode{}
