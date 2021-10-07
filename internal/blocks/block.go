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
)

// Block allows data to be read from directly from file segments.
type Block struct {
	Segments []*FileSegment
	Size     int64
}

func (b *Block) Read(p []byte) (n int, err error) {
	if int64(len(p)) < b.Size {
		return 0, fmt.Errorf("insufficient size")
	}

	var offset int
	for _, segment := range b.Segments {
		n, err := segment.Read(p[offset : int64(offset)+segment.Size])
		offset += n

		if err != nil && err != io.EOF {
			return offset, err
		}
	}

	return offset, nil
}

var _ io.Reader = &Block{}
