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

package storage

import (
	"fmt"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	"github.com/mjpitz/aetherfs/internal/storage/s3"
)

type Config struct {
	Driver string `json:"driver" usage:"configure how information is stored"`

	S3 s3.Config `json:"s3"`
}

type Stores struct {
	BlockAPIServer   blockv1.BlockAPIServer
	DatasetAPIServer datasetv1.DatasetAPIServer
}

func ObtainStores(cfg Config) (*Stores, error) {
	var blockAPI blockv1.BlockAPIServer
	var datasetAPI datasetv1.DatasetAPIServer
	var err error

	switch cfg.Driver {
	case "s3":
		blockAPI, datasetAPI, err = s3.ObtainStores(cfg.S3)

	default:
		err = fmt.Errorf("invalid driver: %s", cfg.Driver)
	}

	if err != nil {
		return nil, err
	}

	return &Stores{
		BlockAPIServer:   blockAPI,
		DatasetAPIServer: datasetAPI,
	}, nil
}
