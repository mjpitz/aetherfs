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

package fs

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
