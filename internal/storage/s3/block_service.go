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
	"context"
	"io"
	"math"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/minio/minio-go/v7"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	"github.com/mjpitz/aetherfs/internal/blocks"
	"github.com/mjpitz/aetherfs/internal/headers"
)

type blockService struct {
	blockv1.UnsafeBlockAPIServer

	s3Client   *minio.Client
	bucketName string
}

func (b *blockService) Lookup(ctx context.Context, request *blockv1.LookupRequest) (*blockv1.LookupResponse, error) {
	objectKey := "blocks/" + request.Signature[0:2] + "/" + request.Signature[2:]

	_, err := b.s3Client.StatObject(ctx, b.bucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		if cast, ok := err.(minio.ErrorResponse); ok {
			switch cast.StatusCode {
			case http.StatusNotFound:
				return nil, status.Errorf(codes.NotFound, "not found")
			}
		}

		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return &blockv1.LookupResponse{}, nil
}

func (b *blockService) Download(request *blockv1.DownloadRequest, call blockv1.BlockAPI_DownloadServer) error {
	logger := ctxzap.Extract(call.Context())

	objectKey := "blocks/" + request.Signature[0:2] + "/" + request.Signature[2:]

	resp, err := b.s3Client.GetObject(call.Context(), b.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		logger.Error("failed to get object", zap.Error(err))
		return status.Errorf(codes.Internal, "internal server error")
	}

	_, err = resp.Seek(request.Offset, io.SeekStart)
	if err != nil {
		logger.Error("seek failed", zap.Error(err))
		return status.Errorf(codes.Internal, "internal server error")
	}

	// read 64KB blocks from resp until request.Size || the remaining file is read is read
	// this should be the same same cache size
	remaining := request.Size

	part := make([]byte, 0, blocks.PartSize)
	for remaining > 0 {
		length := int(math.Min(float64(blocks.PartSize), float64(remaining)))

		_, err := resp.Read(part[:length])
		if err != nil && err != io.EOF {
			logger.Error("read failed", zap.Error(err))
			return status.Errorf(codes.Internal, "internal server error")
		}

		err = call.Send(&blockv1.DownloadResponse{
			Part: part[:length],
		})
		if err != nil {
			logger.Error("send failed", zap.Error(err))
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
		ctxzap.Extract(ctx).Error("failed to put object", zap.Error(err))
		return status.Errorf(codes.Internal, "internal server error")
	}

	return nil
}

var _ blockv1.BlockAPIServer = &blockService{}
