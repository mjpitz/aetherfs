package daemons

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	fsv1 "github.com/mjpitz/aetherfs/api/aetherfs/fs/v1"
)

type fileServerService struct {
	fsv1.UnsafeFileServerAPIServer
}

func (f *fileServerService) Lookup(ctx context.Context, request *fsv1.LookupRequest) (*fsv1.LookupResponse, error) {
	// parse path
	request.GetPath()

	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

var _ fsv1.FileServerAPIServer = &fileServerService{}
