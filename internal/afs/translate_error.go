// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package afs

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
