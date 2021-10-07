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

package blocks

import (
	"fmt"
	"io"
	"os"
)

func closeFile(closer io.Closer) {
	if closer != nil {
		_ = closer.Close()
	}
}

// FileSegment defines a part of a file used to construct a block.
type FileSegment struct {
	FilePath string
	Offset   int64
	Size     int64
}

func (f *FileSegment) Read(p []byte) (n int, err error) {
	if int64(len(p)) < f.Size {
		return 0, fmt.Errorf("insufficient size")
	}

	file, err := os.Open(f.FilePath)
	defer closeFile(file)

	if err != nil {
		return 0, err
	}

	seek, err := file.Seek(f.Offset, io.SeekStart)
	if err != nil {
		return 0, err
	} else if seek != f.Offset {
		return 0, fmt.Errorf("failed to advance")
	}

	return file.Read(p[:f.Size])
}

var _ io.Reader = &FileSegment{}
