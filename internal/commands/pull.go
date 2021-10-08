// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	"github.com/mjpitz/aetherfs/internal/components"
	"github.com/mjpitz/aetherfs/internal/flagset"
	"github.com/mjpitz/aetherfs/internal/fs"
)

const (
	filePermissions os.FileMode = 0644
	dirPermissions  os.FileMode = 0755
)

// PullConfig encapsulates all the configuration required to pull datasets from AetherFS.
type PullConfig struct {
}

// Pull returns a command that downloads datasets from upstream servers
func Pull() *cli.Command {
	cfg := &PullConfig{}

	return &cli.Command{
		Name:  "pull",
		Usage: "Pulls a dataset from AetherFS",
		UsageText: ExampleString(
			"aetherfs pull [options] <path> [dataset...]",
			"aetherfs pull /var/datasets maxmind:v1 private.company.io/maxmind:v2",
			"aetherfs pull -c path/to/application.afs.yaml /var/datasets",
		),
		Flags: flagset.Extract(cfg),
		Action: func(ctx *cli.Context) error {
			logger := ctxzap.Extract(ctx.Context)

			args := ctx.Args().Slice()
			switch len(args) {
			case 0:
				return fmt.Errorf("missing required path")
			case 1:
				return fmt.Errorf("missing datasets")
			}

			datasets := args[1:]
			path, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			_ = os.MkdirAll(path, dirPermissions)
			{
				info, err := os.Stat(path)
				switch {
				case err != nil:
					return errors.Wrapf(err, "failed to make path")
				case !info.IsDir():
					return fmt.Errorf("path is not a directory")
				}
			}

			aetherFSDir := filepath.Join(path, ".aetherfs")
			_ = os.MkdirAll(aetherFSDir, dirPermissions)

			{
				info, err := os.Stat(aetherFSDir)
				switch {
				case err != nil:
					return errors.Wrapf(err, "failed to make aetherfs dir")
				case !info.IsDir():
					return fmt.Errorf(".aetherfs is a file")
				}
			}

			conn, err := components.GRPCClient(ctx.Context, components.GRPCClientConfig{
				Target: lookupHost(),
			})
			if err != nil {
				return err
			}
			defer conn.Close()

			datasetAPI := datasetv1.NewDatasetAPIClient(conn)
			blockAPI := blockv1.NewBlockAPIClient(conn)

			tags := make([]*datasetv1.Tag, 0, len(datasets))
			snapshots := make([]*datasetv1.LookupResponse, 0, len(datasets))

			for _, dataset := range datasets {
				parts := strings.Split(dataset, ":")
				if len(parts) < 2 {
					parts = append(parts, "latest")
				}

				req := &datasetv1.LookupRequest{
					Tag: &datasetv1.Tag{
						Name:    parts[0],
						Version: parts[1],
					},
				}

				resp, err := datasetAPI.Lookup(ctx.Context, req)
				if err != nil {
					return err
				}

				tags = append(tags, req.Tag)
				snapshots = append(snapshots, resp)
			}

			// save snapshots
			for i, snapshot := range snapshots {
				tag := tags[i]

				metadataFile := tag.Name + "." + tag.Version + ".snapshot.afs.json"
				metadataFile = filepath.Join(aetherFSDir, metadataFile)

				datasetDir := tag.Name + "." + tag.Version
				datasetDir = filepath.Join(path, datasetDir)

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

					datasetFile := &fs.DatasetFile{
						Context:     ctx.Context,
						BlockAPI:    blockAPI,
						Dataset:     snapshot.Dataset,
						CurrentPath: file.Name,
						File:        file,
					}

					data := make([]byte, file.Size)
					n, err := datasetFile.Read(data)
					if err != nil {
						return errors.Wrap(err, "failed to download file")
					}

					err = ioutil.WriteFile(filePath, data[:n], filePermissions)
					if err != nil {
						return errors.Wrap(err, "failed to write file")
					}
				}

				// save snapshot

				data, err := json.MarshalIndent(snapshot, "", "  ")
				if err != nil {
					return err
				}

				err = ioutil.WriteFile(metadataFile, data, filePermissions)
				if err != nil {
					return errors.Wrap(err, "failed to write metadata file")
				}
			}

			return nil
		},
		HideHelpCommand: true,
	}
}
