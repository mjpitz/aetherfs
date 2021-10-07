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

package commands

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	"github.com/mjpitz/aetherfs/internal/blocks"
	"github.com/mjpitz/aetherfs/internal/components"
	"github.com/mjpitz/aetherfs/internal/flagset"
	"github.com/mjpitz/aetherfs/internal/headers"
)

// PushConfig encapsulates all the configuration required to push datasets to AetherFS.
type PushConfig struct {
	BlockSize int32 `json:"block_size"   usage:"the maximum number of bytes per block in MiB"`
}

// Push returns a command used to push datasets to upstream servers.
func Push() *cli.Command {
	cfg := &PushConfig{
		BlockSize: 256,
	}

	tags := cli.NewStringSlice() // can't put this in config struct quite yet

	return &cli.Command{
		Name:  "push",
		Usage: "Pushes a dataset into AetherFS",
		UsageText: ExampleString(
			"aetherfs push [options] <path>",
			"aetherfs push -t maxmind:v1 -t private.company.io/maxmind:v2 /tmp/maxmind",
		),
		Flags: append(
			flagset.Extract(cfg),
			[]cli.Flag{
				&cli.StringSliceFlag{
					Name:        "tag",
					Aliases:     []string{"t"},
					Usage:       "name and tag of the dataset being pushed",
					Value:       tags,
					Destination: tags,
					Required:    true,
				},
			}...,
		),
		Action: func(ctx *cli.Context) error {
			logger := ctxzap.Extract(ctx.Context)

			root := ctx.Args().Get(0)
			if root == "" {
				return fmt.Errorf("missing path argument")
			}

			root, err := filepath.Abs(root)
			if err != nil {
				return err
			}

			conn, err := components.GRPCClient(ctx.Context, components.GRPCClientConfig{
				Target: lookupHost(),
			})
			if err != nil {
				return err
			}
			defer conn.Close()

			blockAPI := blockv1.NewBlockAPIClient(conn)
			datasetAPI := datasetv1.NewDatasetAPIClient(conn)

			// cache some metadata for later on to make things easier
			publishRequest := &datasetv1.PublishRequest{
				Dataset: &datasetv1.Dataset{
					BlockSize: cfg.BlockSize * int32(blocks.Mebibyte),
				},
			}

			for _, tag := range tags.Value() {
				parts := strings.Split(tag, ":")
				if len(parts) < 2 {
					parts = append(parts, "latest")
				}

				publishRequest.Tags = append(publishRequest.Tags, &datasetv1.Tag{
					Name:    parts[0],
					Version: parts[1],
				})
			}

			// create a block table to detail which file segments belong to which block.
			// this _should_ allow for concurrent uploads.
			var allBlocks []*blocks.Block
			current := &blocks.Block{}

			err = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
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

			if err != nil {
				return err
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
					return err
				}

				signature, err := blocks.ComputeSignature("sha256", data[:block.Size])
				if err != nil {
					return err
				}

				publishRequest.Dataset.Blocks = append(publishRequest.Dataset.Blocks, signature)
				logger.Info("uploading block", zap.String("signature", signature))

				// attempt to upload
				// the server will reply with an error if the block already exists

				uploadContext := metadata.AppendToOutgoingContext(ctx.Context,
					headers.AetherFSBlockSignature, signature,
					headers.AetherFSBlockSize, strconv.FormatInt(block.Size, 10),
				)

				call, err := blockAPI.Upload(uploadContext)
				if err != nil {
					st, ok := status.FromError(err)
					if ok && st.Code() == codes.AlreadyExists {
						logger.Info("block already exists", zap.String("signature", signature))
						continue BlockLoop
					}
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

					if err != nil {
						st, ok := status.FromError(err)
						if ok && st.Code() == codes.AlreadyExists {
							logger.Info("block already exists", zap.String("signature", signature))
							continue BlockLoop
						}
						return err
					}
				}

				_, err = call.CloseAndRecv()
				if err == io.EOF {
					continue
				} else if err != nil {
					st, ok := status.FromError(err)
					if ok && st.Code() == codes.AlreadyExists {
						logger.Info("block already exists", zap.String("signature", signature))
						continue BlockLoop
					}

					return err
				}
			}

			logger.Info("publishing dataset with tags", zap.Strings("tags", tags.Value()))

			_, err = datasetAPI.Publish(ctx.Context, publishRequest)
			return err
		},
		HideHelpCommand: true,
	}
}
