// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

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
