package daemons

import (
	"context"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	fsv1 "github.com/mjpitz/aetherfs/api/aetherfs/fs/v1"
)

type fileServerService struct {
	fsv1.UnsafeFileServerAPIServer

	blockAPI   blockv1.BlockAPIClient
	datasetAPI datasetv1.DatasetAPIClient
}

func (f *fileServerService) renderFile(ctx context.Context, dataset *datasetv1.Dataset, filePath string) (*fsv1.LookupResponse, error) {
	offset := uint64(0)
	var file *datasetv1.File

	for _, f := range dataset.GetFiles() {
		if f.Name == filePath {
			file = f
			break
		}

		offset += f.GetSize()
	}

	// file should always be present
	// todo: support HTTP range requests

	blocks := dataset.GetBlocks()
	blocksToRead := (file.GetSize() / uint64(dataset.BlockSize)) + 1

	startingBlock := offset / uint64(dataset.BlockSize)
	blockOffset := offset % uint64(dataset.BlockSize)

	for i := uint64(0); i < blocksToRead; i++ {
		f.blockAPI.Download(ctx, &blockv1.DownloadRequest{
			Signature: blocks[startingBlock+i],
			Offset: blockOffset,
		})
	//
	//	resp, err := client.Recv()
	//	switch {
	//	case err == io.EOF:
	//		break
	//	case err != nil:
	//		return nil, err
	//	}
	//
	//	resp.GetPart()
	}

	return &fsv1.LookupResponse{
		Body: "renderFile",
	}, nil
}

func (f *fileServerService) renderDirectory(ctx context.Context, node *FileServerNode) (*fsv1.LookupResponse, error) {
	return &fsv1.LookupResponse{
		Body: "renderDirectory",
	}, nil
}

func (f *fileServerService) Lookup(ctx context.Context, request *fsv1.LookupRequest) (*fsv1.LookupResponse, error) {
	log := ctxzap.Extract(ctx)

	root := new(FileServerNode)
	root.mode = directoryMask

	datasetList, err := f.datasetAPI.List(ctx, &datasetv1.ListRequest{})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Unauthenticated {
			return nil, err
		}

		log.Error("failed to list datasets", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	for _, dataset := range datasetList.GetDatasets() {
		root.insert(dataset, directoryMask&datasetMask)
	}

	var tag *datasetv1.Tag
	var dataset *datasetv1.Dataset

	ptr := root
	parts := strings.Split(request.GetPath(), "/")
	for i, part := range parts {
		val := ptr.children[part]

		if val == nil {
			return nil, status.Errorf(codes.NotFound, "not found")
		}

		if val.IsDataset() {
			if i == len(parts)-1 {
				listTagsResp, err := f.datasetAPI.ListTags(ctx, &datasetv1.ListTagsRequest{
					Name: strings.Join(parts[:i+1], "/"),
				})

				if err != nil {
					st, ok := status.FromError(err)
					if ok && st.Code() == codes.Unauthenticated {
						// translate to not found
						return nil, status.Errorf(codes.NotFound, "not found")
					}

					log.Error("failed to list tags for dataset", zap.Error(err))
					return nil, status.Errorf(codes.Internal, "internal server error")
				}

				for _, tag := range listTagsResp.GetTags() {
					val.insert(tag.Version, directoryMask&tagMask)
				}

				continue // skip the following block
			}

			// peak ahead and load the specific version
			tag = &datasetv1.Tag{
				Name:    strings.Join(parts[:i+1], "/"),
				Version: parts[i+1],
			}

			lookupResp, err := f.datasetAPI.Lookup(ctx, &datasetv1.LookupRequest{
				Tag: tag,
			})

			if err != nil {
				st, ok := status.FromError(err)
				if ok && st.Code() == codes.Unauthenticated {
					return nil, status.Errorf(codes.NotFound, "not found")
				}

				log.Error("failed to lookup dataset", zap.Error(err))
				return nil, status.Errorf(codes.Internal, "internal server error")
			}

			// create a tag node
			val.insert(tag.Version, directoryMask&tagMask)

			dataset = lookupResp.GetDataset()
			for _, file := range dataset.GetFiles() {
				val.children[tag.Version].insert(file.Name, fileMask)
			}
		}

		ptr = val
	}

	switch {
	case ptr.IsFile():
		if tag == nil || dataset == nil {
			// these should be set when were looking at a file
			// files only exist within tagged datasets
			return nil, status.Errorf(codes.Internal, "internal server error")
		}

		filePath := strings.TrimPrefix(request.GetPath(), tag.GetName()+"/"+tag.GetVersion()+"/")

		return f.renderFile(ctx, dataset, filePath)
	case ptr.IsDirectory():
		return f.renderDirectory(ctx, ptr)
	}

	// must be either a file or directory
	// did you forget to set a bit somewhere?
	return nil, status.Errorf(codes.Internal, "internal server error")
}

var _ fsv1.FileServerAPIServer = &fileServerService{}
