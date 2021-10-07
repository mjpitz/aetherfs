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

package fs

import (
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// translateError takes in an arbitrary error and attempts to convert it to a more meaningful error code.
func translateError(err error) error {
	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case codes.Unauthenticated:
			return os.ErrPermission
		case codes.NotFound:
			return os.ErrNotExist
		case codes.DeadlineExceeded:
			return os.ErrDeadlineExceeded
		}
	}

	return err
}
