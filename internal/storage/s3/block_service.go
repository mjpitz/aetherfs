// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package s3

import (
	"context"
	"io"
	"math"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/minio/minio-go/v7"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	"github.com/mjpitz/aetherfs/internal/headers"
)

type blockService struct {
	blockv1.UnsafeBlockAPIServer

	s3Client   *minio.Client
	bucketName string
}

func (b *blockService) Lookup(ctx context.Context, request *blockv1.LookupRequest) (*blockv1.LookupResponse, error) {
	objectKey := "blocks/" + request.Signature[0:2] + "/" + request.Signature[2:]

	info, err := b.s3Client.StatObject(ctx, b.bucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	if info.Key == "" {
		return nil, status.Errorf(codes.NotFound, "not found")
	}

	return &blockv1.LookupResponse{}, nil
}

func (b *blockService) Download(request *blockv1.DownloadRequest, call blockv1.BlockAPI_DownloadServer) error {
	objectKey := "blocks/" + request.Signature[0:2] + "/" + request.Signature[2:]

	resp, err := b.s3Client.GetObject(call.Context(), b.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return status.Errorf(codes.Internal, "internal server error")
	}

	_, err = resp.Seek(request.Offset, io.SeekStart)
	if err != nil {
		return status.Errorf(codes.Internal, "internal server error")
	}

	// read 64KB blocks from resp until request.Size || the remaining file is read is read
	// this should be the same same cache size
	blockSize := (1 << 10) * 64
	remaining := request.Size

	part := make([]byte, 0, blockSize)
	for remaining > 0 {
		length := int(math.Min(float64(blockSize), float64(remaining)))

		_, err := resp.Read(part[:length])
		if err != nil {
			return status.Errorf(codes.Internal, "internal server error")
		}

		err = call.Send(&blockv1.DownloadResponse{
			Part: part[:length],
		})
		if err != nil {
			return status.Errorf(codes.Internal, "")
		}

		remaining -= int64(length)
	}

	return nil
}

func (b *blockService) Upload(call blockv1.BlockAPI_UploadServer) error {
	ctx := call.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	signatures := md.Get(headers.AetherFSBlockSignature)
	sizes := md.Get(headers.AetherFSBlockSize)

	if len(signatures) == 0 || len(sizes) == 0 {
		return status.Errorf(codes.InvalidArgument,
			"missing %s or %s header", headers.AetherFSBlockSignature, headers.AetherFSBlockSize)
	}

	expectedSignature := signatures[0]
	expectedSize, err := strconv.ParseInt(sizes[0], 10, 64)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "%s is not a number", headers.AetherFSBlockSize)
	}

	_, err = b.Lookup(ctx, &blockv1.LookupRequest{
		Signature: expectedSignature,
	})
	st, ok := status.FromError(err)

	switch {
	case err == nil:
		return status.Errorf(codes.AlreadyExists, "already exists")
	case ok && st.Code() == codes.Internal:
		return err
	}

	reader := &uploadReader{call: call}
	// todo: add checksum verification
	// if the computed checksum does not match the provided checksum, we should error and fail the put

	objectKey := "blocks/" + expectedSignature[0:2] + "/" + expectedSignature[2:]

	_, err = b.s3Client.PutObject(ctx, b.bucketName, objectKey, reader, expectedSize, minio.PutObjectOptions{})
	if err != nil {
		return status.Errorf(codes.Internal, "internal server error")
	}

	return nil
}

var _ blockv1.BlockAPIServer = &blockService{}
