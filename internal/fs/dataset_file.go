// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package fs

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
)

type DatasetFile struct {
	Context context.Context

	BlockAPI blockv1.BlockAPIClient

	Dataset     *datasetv1.Dataset
	CurrentPath string
	File        *datasetv1.File

	fileOffset int64
}

func (f *DatasetFile) Close() error {
	return nil
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func (f *DatasetFile) Read(p []byte) (n int, err error) {
	if f.File == nil {
		return 0, os.ErrInvalid
	}

	if f.fileOffset >= f.File.Size {
		return 0, io.EOF
	}

	blockSize := int64(f.Dataset.BlockSize)
	fileOffset := f.fileOffset

	// factor in fileOffset which can reduce the total number of bytes that can be read
	numBytesToRead := min(int64(len(p)), f.File.Size-fileOffset)

	var numBlocksToRead int64
	if numBytesToRead%blockSize > 0 {
		numBlocksToRead = 1
	}
	numBlocksToRead += numBytesToRead / blockSize

	var datasetFileOffset int64
	for _, file := range f.Dataset.Files {
		if file.Name == f.File.Name {
			break
		}

		datasetFileOffset += file.Size
	}

	// factor in fileOffset as it impacts where we start reading data
	readOffset := datasetFileOffset + fileOffset

	startingBlock := readOffset / blockSize
	blockOffset := readOffset % blockSize

	bytesRead := 0
	for i := startingBlock; i < startingBlock+numBlocksToRead; i++ {
		stream, err := f.BlockAPI.Download(f.Context, &blockv1.DownloadRequest{
			Signature: f.Dataset.Blocks[i],
			Offset:    blockOffset,
			Size:      min(blockSize, numBytesToRead-int64(bytesRead)),
		})
		if err != nil {
			return bytesRead, translateError(err)
		}

		var resp *blockv1.DownloadResponse

	LOOP:
		for {
			resp, err = stream.Recv()
			copy(p[bytesRead:], resp.GetPart())
			bytesRead += len(resp.GetPart())

			switch {
			case err == io.EOF:
				break LOOP
			case err != nil:
				// translate err
				return bytesRead, translateError(err)
			}
		}

		// every subsequent block should be read from the start
		blockOffset = 0
	}

	if bytesRead < len(p) {
		err = io.EOF
	}

	return bytesRead, err
}

func (f *DatasetFile) Seek(offset int64, whence int) (int64, error) {
	if f.File == nil {
		return 0, os.ErrInvalid
	}

	var next int64
	switch whence {
	case io.SeekStart:
		next = offset
	case io.SeekCurrent:
		next = f.fileOffset + offset
	case io.SeekEnd:
		next = f.File.Size + offset
	default:
		return 0, errors.New("daemons.DatasetFile.Seek: invalid whence")
	}

	if next < 0 {
		return 0, errors.New("daemons.DatasetFile.Seek: negative position")
	}

	f.fileOffset = next
	return next, nil
}

func (f *DatasetFile) Readdir(count int) ([]fs.FileInfo, error) {
	seen := make(map[string]bool)
	var infos []fs.FileInfo

	prefix := strings.TrimSuffix(f.CurrentPath, "/")
	if prefix != "" {
		prefix = prefix + "/"
	}

	for _, file := range f.Dataset.GetFiles() {
		if strings.HasPrefix(file.Name, prefix) {
			remaining := strings.TrimPrefix(file.Name, prefix)
			remaining = strings.TrimPrefix(remaining, "/")

			idx := strings.Index(remaining, "/")

			switch {
			case idx == -1:
				infos = append(infos, &fileInfo{
					name: remaining,
					file: file,
				})
			case !seen[remaining[:idx]]:
				infos = append(infos, &fileInfo{
					name: remaining[:idx],
				})
				seen[remaining[:idx]] = true
			}
		}
	}

	return infos, nil
}

func (f *DatasetFile) Stat() (fs.FileInfo, error) {
	name := f.CurrentPath[strings.LastIndex(f.CurrentPath, "/")+1:]

	return &fileInfo{
		name: name,
		file: f.File,
	}, nil
}

var _ http.File = &DatasetFile{}
