// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package proxy

import (
	"context"
	"io"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
)

type blockService struct {
	blockv1.UnsafeBlockAPIServer

	delegate blockv1.BlockAPIClient
}

func (b *blockService) Lookup(ctx context.Context, request *blockv1.LookupRequest) (*blockv1.LookupResponse, error) {
	return b.delegate.Lookup(ctx, request)
}

func (b *blockService) Download(request *blockv1.DownloadRequest, call blockv1.BlockAPI_DownloadServer) error {
	up, err := b.delegate.Download(call.Context(), request)
	if err != nil {
		return err
	}

	msg := &blockv1.DownloadResponse{}
	for {
		err := up.RecvMsg(msg)
		if err != nil {
			return err
		}

		err = call.Send(msg)
		if err != nil {
			return err
		}
	}
}

func (b *blockService) Upload(call blockv1.BlockAPI_UploadServer) error {
	up, err := b.delegate.Upload(call.Context())
	if err != nil {
		return err
	}

	msg := &blockv1.UploadRequest{}
	for {
		err = call.RecvMsg(msg)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		err = up.Send(msg)
		if err != nil {
			return err
		}
	}

	resp, err := up.CloseAndRecv()
	if err != nil {
		return err
	}

	return call.SendAndClose(resp)
}

var _ blockv1.BlockAPIServer = &blockService{}
