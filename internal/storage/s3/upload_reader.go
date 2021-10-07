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
