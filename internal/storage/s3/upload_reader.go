// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package s3

import (
	"io"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
)

type uploadReader struct {
	buffer []byte
	call   blockv1.BlockAPI_UploadServer
}

func (r *uploadReader) Read(p []byte) (n int, err error) {
	for len(p) > len(r.buffer) {
		req, err := r.call.Recv()
		if err != nil {
			return 0, err
		}

		r.buffer = append(r.buffer, req.Part...)
	}

	n = copy(p, r.buffer[:len(p)])
	r.buffer = r.buffer[len(p):]

	return n, nil
}

var _ io.Reader = &uploadReader{}
