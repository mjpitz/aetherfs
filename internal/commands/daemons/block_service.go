// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package daemons

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
)

type blockService struct {
	blockv1.UnsafeBlockAPIServer
}

func (b *blockService) Lookup(ctx context.Context, request *blockv1.LookupRequest) (*blockv1.LookupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (b *blockService) Download(request *blockv1.DownloadRequest, server blockv1.BlockAPI_DownloadServer) error {
	return status.Errorf(codes.Unimplemented, "unimplemented")
}

func (b *blockService) Upload(server blockv1.BlockAPI_UploadServer) error {
	return status.Errorf(codes.Unimplemented, "unimplemented")
}

var _ blockv1.BlockAPIServer = &blockService{}
