// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	agentv1 "github.com/mjpitz/aetherfs/api/aetherfs/agent/v1"
	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	"github.com/mjpitz/aetherfs/internal/blocks"
	"github.com/mjpitz/aetherfs/internal/headers"
	"github.com/mjpitz/myago/clocks"
	"github.com/mjpitz/myago/vfs"
)

type Service struct {
	agentv1.UnsafeAgentAPIServer

	InitiateShutdown func()
	BlockAPI         blockv1.BlockAPIClient
	DatasetAPI       datasetv1.DatasetAPIClient

	ongoing  int32
	shutdown int32
}

// this really needs to get broken up into some smaller methods
func (s *Service) publish(ctx context.Context, request *agentv1.PublishRequest) (*agentv1.PublishResponse, error) {
	defer atomic.AddInt32(&s.ongoing, -1)

	// cache some metadata for later on to make things easier
	publishRequest := &datasetv1.PublishRequest{
		Dataset: &datasetv1.Dataset{
			BlockSize: request.BlockSize,
		},
		Tags: request.Tags,
	}

	// create a block table to detail which file segments belong to which block.
	// this _should_ allow for concurrent uploads.
	var allBlocks []*blocks.Block
	current := &blocks.Block{}

	logger := ctxzap.Extract(ctx)
	vfs := vfs.Extract(ctx)
	root := request.GetPath()

	err := afero.Walk(vfs, root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip non-regular files for now
		if !info.Mode().IsRegular() {
			return nil
		}

		// store some local metadata
		file := &datasetv1.File{
			Name:         strings.TrimPrefix(strings.TrimPrefix(path, root), "/"),
			Size:         info.Size(),
			LastModified: timestamppb.New(info.ModTime()),
		}
		publishRequest.Dataset.Files = append(publishRequest.Dataset.Files, file)

		// break large files up into multiple blocks
		// glob small files into single block
		remainingInFile := file.Size
		offset := int64(0)

		for remainingInFile > 0 {
			// how many bytes to grab
			size := int64(publishRequest.Dataset.BlockSize) - current.Size
			if remainingInFile < size {
				size = remainingInFile
			}

			// update block table
			current.Segments = append(current.Segments, &blocks.FileSegment{
				FilePath: path,
				Offset:   offset,
				Size:     size,
			})
			current.Size += size

			// advance pointer and decrement step
			offset += size
			remainingInFile -= size

			switch {
			case current.Size > int64(publishRequest.Dataset.BlockSize):
				// pebcak - programmer error
				return fmt.Errorf("block overflow")

			case current.Size == int64(publishRequest.Dataset.BlockSize):
				// roll over full blocks
				allBlocks = append(allBlocks, current)
				current = &blocks.Block{}
			}
		}

		return nil
	})

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, status.Errorf(codes.InvalidArgument, "associated file path does not exist")
	case err != nil:
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	// catch any partial blocks
	if current.Size > 0 {
		allBlocks = append(allBlocks, current)
	}

	// keep memory usage low and reduce garbage collection by re-using byte block
	data := make([]byte, publishRequest.Dataset.BlockSize)

BlockLoop:
	for _, block := range allBlocks {
		_, err := block.Read(data[:block.Size])
		if err != nil && err != io.EOF {
			return nil, status.Errorf(codes.Internal, err.Error())
		}

		signature, err := blocks.ComputeSignature("sha256", data[:block.Size])
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}

		publishRequest.Dataset.Blocks = append(publishRequest.Dataset.Blocks, signature)
		logger.Info("uploading block", zap.String("signature", signature))

		// attempt to upload
		// the server will reply with an error if the block already exists

		uploadContext := metadata.AppendToOutgoingContext(ctx,
			headers.AetherFSBlockSignature, signature,
			headers.AetherFSBlockSize, strconv.FormatInt(block.Size, 10),
		)

		call, err := s.BlockAPI.Upload(uploadContext)

		st, ok := status.FromError(err)
		if err == io.EOF || (ok && st.Code() == codes.AlreadyExists) {
			logger.Info("block already exists", zap.String("signature", signature))
			continue BlockLoop
		} else if err != nil {
			return nil, err
		}

		for i := int64(0); i < block.Size; i += int64(blocks.PartSize) {
			end := i + int64(blocks.PartSize)
			if end > block.Size {
				end = block.Size
			}

			err = call.Send(&blockv1.UploadRequest{
				Part: data[i:end],
			})

			st, ok := status.FromError(err)
			if err == io.EOF || (ok && st.Code() == codes.AlreadyExists) {
				logger.Info("block already exists", zap.String("signature", signature))

				continue BlockLoop
			} else if err != nil {
				return nil, err
			}
		}

		_, err = call.CloseAndRecv()
		if err == io.EOF {
			continue BlockLoop
		} else if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.AlreadyExists {
				logger.Info("block already exists", zap.String("signature", signature))

				continue BlockLoop
			}

			return nil, err
		}
	}

	logger.Info("publishing dataset with tags")
	_, err = s.DatasetAPI.Publish(ctx, publishRequest)

	return nil, err
}

func (s *Service) Publish(ctx context.Context, request *agentv1.PublishRequest) (*agentv1.PublishResponse, error) {
	if atomic.LoadInt32(&s.shutdown) > 0 {
		return nil, status.Error(codes.InvalidArgument, "shutdown already initiated")
	}

	atomic.AddInt32(&s.ongoing, 1)

	if request.Sync {
		return s.publish(ctx, request)
	}

	go func() {
		_, err := s.publish(ctx, request)
		if err != nil {
			ctxzap.Extract(ctx).Error("failed to publish dataset", zap.Error(err))
		}
	}()

	return &agentv1.PublishResponse{}, nil
}

func (s *Service) Subscribe(server agentv1.AgentAPI_SubscribeServer) error {
	return status.Error(codes.Unimplemented, "unimplemented")
}

func (s *Service) GracefulShutdown(ctx context.Context, request *agentv1.GracefulShutdownRequest) (*agentv1.GracefulShutdownResponse, error) {
	if s.InitiateShutdown == nil {
		return nil, status.Errorf(codes.Unimplemented, "unimplemented")
	}

	if !atomic.CompareAndSwapInt32(&s.shutdown, 0, 1) {
		return nil, status.Error(codes.InvalidArgument, "shutdown already initiated")
	}

	clock := clocks.Extract(ctx)

	ticker := clock.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.Chan():
			if v := atomic.LoadInt32(&s.ongoing); v == 0 {
				go s.InitiateShutdown()

				return &agentv1.GracefulShutdownResponse{}, nil
			}

			ticker.Stop()
			ticker = clock.NewTicker(time.Second)
		}
	}
}

var _ agentv1.AgentAPIServer = &Service{}
