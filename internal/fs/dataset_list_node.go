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
