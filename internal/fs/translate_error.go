// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021
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
