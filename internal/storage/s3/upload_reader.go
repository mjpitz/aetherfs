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
	call   blockv1.BlockAPI_UploadServer
	buffer []byte
	done   bool
}

func (r *uploadReader) Read(p []byte) (n int, err error) {
	for !r.done && len(p) > len(r.buffer) {
		req, err := r.call.Recv()
		r.buffer = append(r.buffer, req.GetPart()...)

		if err == io.EOF {
			r.done = true
			break
		} else if err != nil {
			r.done = true
			return 0, err
		}
	}

	n = len(p)
	if n > len(r.buffer) {
		n = len(r.buffer)
		err = io.EOF
	}

	n = copy(p, r.buffer[:n])
	r.buffer = r.buffer[n:]

	return n, err
}

var _ io.Reader = &uploadReader{}
