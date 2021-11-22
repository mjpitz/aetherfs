// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	agentv1 "github.com/mjpitz/aetherfs/api/aetherfs/agent/v1"
	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	"github.com/mjpitz/aetherfs/internal/blocks"
	"github.com/mjpitz/aetherfs/internal/components"
	"github.com/mjpitz/aetherfs/internal/dataset"
	afs "github.com/mjpitz/aetherfs/internal/fs"
	"github.com/mjpitz/aetherfs/internal/headers"
	"github.com/mjpitz/myago/clocks"
	"github.com/mjpitz/myago/vfs"
	"github.com/mjpitz/myago/zaputil"
)

const (
	filePermissions os.FileMode = 0644
	dirPermissions  os.FileMode = 0755
)

type Service struct {
	agentv1.UnsafeAgentAPIServer

	InitiateShutdown func()

	ongoing  int32
	shutdown int32
}

// this really needs to get broken up into some smaller methods
func (s *Service) publish(ctx context.Context, root string, host string, request *datasetv1.PublishRequest) error {
	// TODO: translate host to credentials

	conn, err := components.GRPCClient(ctx, components.GRPCClientConfig{
		Target: host,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	blockAPI := blockv1.NewBlockAPIClient(conn)
	datasetAPI := datasetv1.NewDatasetAPIClient(conn)

	// create a block table to detail which file segments belong to which block.
	// this _should_ allow for concurrent uploads.
	var allBlocks []*blocks.Block
	current := &blocks.Block{}

	logger := ctxzap.Extract(ctx).With(zap.String("host", host))

	err = afero.Walk(vfs.Extract(ctx), root, func(path string, info fs.FileInfo, err error) error {
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
		request.Dataset.Files = append(request.Dataset.Files, file)

		// break large files up into multiple blocks
		// glob small files into single block
		remainingInFile := file.Size
		offset := int64(0)

		for remainingInFile > 0 {
			// how many bytes to grab
			size := int64(request.Dataset.BlockSize) - current.Size
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
			case current.Size > int64(request.Dataset.BlockSize):
				// pebcak - programmer error
				return fmt.Errorf("block overflow")

			case current.Size == int64(request.Dataset.BlockSize):
				// roll over full blocks
				allBlocks = append(allBlocks, current)
				current = &blocks.Block{}
			}
		}

		return nil
	})

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return status.Errorf(codes.InvalidArgument, "associated file path does not exist")
	case err != nil:
		return status.Errorf(codes.Internal, err.Error())
	}

	// catch any partial blocks
	if current.Size > 0 {
		allBlocks = append(allBlocks, current)
	}

	// keep memory usage low and reduce garbage collection by re-using byte block
	data := make([]byte, request.Dataset.BlockSize)

BlockLoop:
	for _, block := range allBlocks {
		_, err := block.Read(data[:block.Size])
		if err != nil && err != io.EOF {
			return status.Errorf(codes.Internal, err.Error())
		}

		signature, err := blocks.ComputeSignature("sha256", data[:block.Size])
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}

		request.Dataset.Blocks = append(request.Dataset.Blocks, signature)
		logger.Info("uploading block", zap.String("signature", signature))

		// attempt to upload
		// the server will reply with an error if the block already exists

		uploadContext := metadata.AppendToOutgoingContext(ctx,
			headers.AetherFSBlockSignature, signature,
			headers.AetherFSBlockSize, strconv.FormatInt(block.Size, 10),
		)

		call, err := blockAPI.Upload(uploadContext)

		st, ok := status.FromError(err)
		if err == io.EOF || (ok && st.Code() == codes.AlreadyExists) {
			logger.Info("block already exists", zap.String("signature", signature))
			continue BlockLoop
		} else if err != nil {
			return err
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
				return err
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

			return err
		}
	}

	logger.Info("publishing dataset with tags")
	_, err = datasetAPI.Publish(ctx, request)

	return err
}

func (s *Service) publishAsync(ctx context.Context, request *agentv1.PublishRequest, tagsByHost map[string][]*datasetv1.Tag) (*agentv1.PublishResponse, error) {
	defer atomic.AddInt32(&s.ongoing, -1)

	group, ctx := errgroup.WithContext(ctx)

	publishAsync := func(host string, tags []*datasetv1.Tag) {
		req := &datasetv1.PublishRequest{
			Dataset: &datasetv1.Dataset{
				BlockSize: request.BlockSize,
			},
			Tags: tags,
		}

		group.Go(func() error {
			zaputil.Extract(ctx).Info("running", zap.String("target", host), zap.Stringer("req", req))
			return s.publish(ctx, request.Path, host, req)
		})
	}

	for host, tags := range tagsByHost {
		publishAsync(host, tags)
	}

	err := group.Wait()
	if err != nil {
		return nil, err
	}

	return &agentv1.PublishResponse{}, nil
}

func (s *Service) Publish(ctx context.Context, request *agentv1.PublishRequest) (*agentv1.PublishResponse, error) {
	tagsByHost := make(map[string][]*datasetv1.Tag)
	for _, tag := range request.Tags {
		t := &dataset.Tag{}
		err := t.UnmarshalText([]byte(tag))
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid tag %s", tag)
		}

		tagsByHost[t.Host] = append(tagsByHost[t.Host], &datasetv1.Tag{
			Name:    t.Dataset,
			Version: t.Version,
		})
	}

	if atomic.LoadInt32(&s.shutdown) > 0 {
		return nil, status.Error(codes.InvalidArgument, "shutdown already initiated")
	}

	atomic.AddInt32(&s.ongoing, 1)

	if request.Sync {
		return s.publishAsync(ctx, request, tagsByHost)
	}

	go func() {
		_, err := s.publishAsync(ctx, request, tagsByHost)
		if err != nil {
			ctxzap.Extract(ctx).Error("failed to publish dataset", zap.Error(err))
		}
	}()

	return &agentv1.PublishResponse{}, nil
}

func (s *Service) subscribe(ctx context.Context, host string, tags []*datasetv1.Tag, aetherFSDir string, resp *agentv1.SubscribeResponse) error {
	// TODO: translate host to credentials

	conn, err := components.GRPCClient(ctx, components.GRPCClientConfig{
		Target: host,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	logger := ctxzap.Extract(ctx).With(zap.String("host", host))

	blockAPI := blockv1.NewBlockAPIClient(conn)
	datasetAPI := datasetv1.NewDatasetAPIClient(conn)

	snapshots := make([]*datasetv1.LookupResponse, 0, len(tags))

	for _, tag := range tags {
		req := &datasetv1.LookupRequest{
			Tag: tag,
		}

		resp, err := datasetAPI.Lookup(ctx, req)
		if err != nil {
			return err
		}

		snapshots = append(snapshots, resp)
	}

	// save snapshots
	for i, snapshot := range snapshots {
		tag := tags[i]

		metadataFile := tag.Name + "." + tag.Version + ".snapshot.afs.json"
		metadataFile = filepath.Join(aetherFSDir, metadataFile)

		datasetDir := resp.Paths[host+"/"+tag.Name+":"+tag.Version]

		_, err := os.Stat(metadataFile)
		if err == nil {
			continue
		}

		// download files
		// this could definitely be done in a more efficient way, but this is a good start

		logger.Info("downloading dataset", zap.String("name", tag.Name), zap.String("tag", tag.Version))

		_ = os.MkdirAll(datasetDir, dirPermissions)
		for _, file := range snapshot.Dataset.Files {
			filePath := filepath.Join(datasetDir, file.Name)
			fileDir := filepath.Dir(filePath)

			_ = os.MkdirAll(fileDir, dirPermissions)

			logger.Info("downloading file", zap.String("file", file.Name))

			datasetFile := &afs.DatasetFile{
				Context:     ctx,
				BlockAPI:    blockAPI,
				Dataset:     snapshot.Dataset,
				CurrentPath: file.Name,
				File:        file,
			}

			data := make([]byte, file.Size)
			if file.Size > 0 {
				n, err := datasetFile.Read(data)
				if err != nil {
					return status.Errorf(codes.Internal, "failed to download file")
				}
				data = data[:n]
			}

			err = ioutil.WriteFile(filePath, data[:], filePermissions)
			if err != nil {
				return status.Errorf(codes.Internal, "failed to write file")
			}
		}

		// save snapshot

		opts := protojson.MarshalOptions{
			Multiline: true,
			Indent:    "  ",
		}

		data, err := opts.Marshal(snapshot)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(metadataFile, data, filePermissions)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to write metadata file")
		}
	}

	return nil
}

func (s *Service) subscribeAsync(ctx context.Context, tagsByHost map[string][]*datasetv1.Tag, aetherFSDir string, resp *agentv1.SubscribeResponse) error {
	defer atomic.AddInt32(&s.ongoing, -1)

	group, ctx := errgroup.WithContext(ctx)

	subscribeAsync := func(host string, tags []*datasetv1.Tag) {
		group.Go(func() error {
			return s.subscribe(ctx, host, tags, aetherFSDir, resp)
		})
	}

	for host, tags := range tagsByHost {
		subscribeAsync(host, tags)
	}

	return group.Wait()
}

func (s *Service) Subscribe(ctx context.Context, request *agentv1.SubscribeRequest) (*agentv1.SubscribeResponse, error) {
	if len(request.Path) == 0 {
		request.Path = afero.GetTempDir(vfs.Extract(ctx), "aetherfs")
	}

	resp := &agentv1.SubscribeResponse{
		Paths: make(map[string]string),
	}

	tagsByHost := make(map[string][]*datasetv1.Tag)
	for _, tag := range request.Tags {
		t := &dataset.Tag{}
		err := t.UnmarshalText([]byte(tag))
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid tag %s", tag)
		}

		tagsByHost[t.Host] = append(tagsByHost[t.Host], &datasetv1.Tag{
			Name:    t.Dataset,
			Version: t.Version,
		})

		resp.Paths[t.String()] = filepath.Join(request.Path, t.Dataset, t.Version)
	}

	if atomic.LoadInt32(&s.shutdown) > 0 {
		return nil, status.Error(codes.InvalidArgument, "shutdown already initiated")
	}

	atomic.AddInt32(&s.ongoing, 1)

	_ = os.MkdirAll(request.Path, 0755)
	{
		info, err := os.Stat(request.Path)
		switch {
		case err != nil:
			return nil, status.Error(codes.InvalidArgument, "failed to make directory")
		case !info.IsDir():
			return nil, status.Error(codes.InvalidArgument, "path is not a directory")
		}
	}

	aetherFSDir := filepath.Join(request.Path, ".aetherfs")
	_ = os.MkdirAll(aetherFSDir, 0755)

	{
		info, err := os.Stat(aetherFSDir)
		switch {
		case err != nil:
			return nil, status.Error(codes.InvalidArgument, "failed to make aetherfs directory")
		case !info.IsDir():
			return nil, status.Error(codes.InvalidArgument, ".aetherfs is a file")
		}
	}

	var err error
	if request.Sync {
		err = s.subscribeAsync(ctx, tagsByHost, aetherFSDir, resp)

	} else {
		go func() {
			err := s.subscribeAsync(ctx, tagsByHost, aetherFSDir, resp)
			if err != nil {
				ctxzap.Extract(ctx).Error("failed to publish dataset", zap.Error(err))
			}
		}()
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *Service) GracefulShutdown(ctx context.Context, _ *agentv1.GracefulShutdownRequest) (*agentv1.GracefulShutdownResponse, error) {
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

func (s *Service) WatchSubscription(_ agentv1.AgentAPI_WatchSubscriptionServer) error {
	return status.Errorf(codes.Unimplemented, "unimplemented")
}

var _ agentv1.AgentAPIServer = &Service{}
