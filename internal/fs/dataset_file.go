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

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func (f *datasetFile) Read(p []byte) (n int, err error) {
	if uint64(f.fileOffset) >= f.file.Size {
		return 0, io.EOF
	}

	blockSize := uint64(f.dataset.BlockSize)
	fileOffset := uint64(f.fileOffset)

	// factor in fileOffset which can reduce the total number of bytes that can be read
	numBytesToRead := min(uint64(len(p)), f.file.Size-fileOffset)

	var numBlocksToRead uint64
	if numBytesToRead%blockSize > 0 {
		numBlocksToRead = 1
	}
	numBlocksToRead += numBytesToRead / blockSize

	var datasetFileOffset uint64
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

	pi := 0
	for i := startingBlock; i < startingBlock+numBlocksToRead; i++ {
		stream, err := f.blockAPI.Download(f.ctx, &blockv1.DownloadRequest{
			Signature: f.dataset.Blocks[i],
			Offset:    blockOffset,
			Size:      min(blockSize, numBytesToRead-uint64(pi)),
		})
		if err != nil {
			return pi, translateError(err)
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
				return pi, translateError(err)
			}

			copy(p[pi:], resp.Part)
			pi += len(resp.Part)
		}

		// every subsequent block should be read from the start
		blockOffset = 0
	}

	if pi < len(p) {
		err = io.EOF
	}

	return pi, err
}

func (f *datasetFile) Seek(offset int64, whence int) (int64, error) {
	var next int64
	switch whence {
	case io.SeekStart:
		next = offset
	case io.SeekCurrent:
		next = f.fileOffset + offset
	case io.SeekEnd:
		next = int64(f.file.Size) + offset
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

	for _, file := range f.dataset.GetFiles() {
		if strings.HasPrefix(file.Name, f.filePath) {
			remaining := strings.TrimPrefix(file.Name, f.filePath)
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
