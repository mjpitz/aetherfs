// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

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

type datasetFile struct {
	ctx context.Context

	blockAPI blockv1.BlockAPIClient

	dataset    *datasetv1.Dataset
	filePath   string
	file       *datasetv1.File
	fileOffset int64
}

func (f *datasetFile) Close() error {
	return nil
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func (f *datasetFile) Read(p []byte) (n int, err error) {
	if f.file == nil {
		return 0, os.ErrInvalid
	}

	if f.fileOffset >= f.file.Size {
		return 0, io.EOF
	}

	blockSize := int64(f.dataset.BlockSize)
	fileOffset := f.fileOffset

	// factor in fileOffset which can reduce the total number of bytes that can be read
	numBytesToRead := min(int64(len(p)), f.file.Size-fileOffset)

	var numBlocksToRead int64
	if numBytesToRead%blockSize > 0 {
		numBlocksToRead = 1
	}
	numBlocksToRead += numBytesToRead / blockSize

	var datasetFileOffset int64
	for _, file := range f.dataset.Files {
		if file.Name == f.file.Name {
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
		stream, err := f.blockAPI.Download(f.ctx, &blockv1.DownloadRequest{
			Signature: f.dataset.Blocks[i],
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
			switch {
			case err == io.EOF:
				break LOOP
			case err != nil:
				// translate err
				return bytesRead, translateError(err)
			}

			copy(p[bytesRead:], resp.Part)
			bytesRead += len(resp.Part)
		}

		// every subsequent block should be read from the start
		blockOffset = 0
	}

	if bytesRead < len(p) {
		err = io.EOF
	}

	return bytesRead, err
}

func (f *datasetFile) Seek(offset int64, whence int) (int64, error) {
	if f.file == nil {
		return 0, os.ErrInvalid
	}

	var next int64
	switch whence {
	case io.SeekStart:
		next = offset
	case io.SeekCurrent:
		next = f.fileOffset + offset
	case io.SeekEnd:
		next = f.file.Size + offset
	default:
		return 0, errors.New("daemons.datasetFile.Seek: invalid whence")
	}

	if next < 0 {
		return 0, errors.New("daemons.datasetFile.Seek: negative position")
	}

	f.fileOffset = next
	return next, nil
}

func (f *datasetFile) Readdir(count int) ([]fs.FileInfo, error) {
	seen := make(map[string]bool)
	var infos []fs.FileInfo

	prefix := strings.TrimSuffix(f.filePath, "/")
	if prefix != "" {
		prefix = prefix + "/"
	}

	for _, file := range f.dataset.GetFiles() {
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

func (f *datasetFile) Stat() (fs.FileInfo, error) {
	name := f.filePath[strings.LastIndex(f.filePath, "/")+1:]

	return &fileInfo{
		name: name,
		file: f.file,
	}, nil
}

var _ http.File = &datasetFile{}
