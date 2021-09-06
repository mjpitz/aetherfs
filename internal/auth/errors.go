package auth

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errUnauthorized = status.Error(codes.Unauthenticated, "unauthorized")
var errInternal = status.Error(codes.Internal, "internal server error")
