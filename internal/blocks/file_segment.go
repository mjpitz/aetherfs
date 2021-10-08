// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

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
