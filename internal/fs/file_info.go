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
