// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package auth

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errUnauthorized = status.Error(codes.Unauthenticated, "unauthorized")
var errInternal = status.Error(codes.Internal, "internal server error")
