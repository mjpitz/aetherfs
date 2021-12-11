// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package components

import (
	"context"
	"fmt"
	"net"

	"github.com/spf13/afero"
	"github.com/willscott/go-nfs"
	"github.com/willscott/go-nfs/helpers"
	"go.uber.org/zap"

	"github.com/mjpitz/aetherfs/internal/afs"
	"github.com/mjpitz/myago/lifecycle"
	"github.com/mjpitz/myago/zaputil"
)

type NFSServerConfig struct {
	Enable bool `json:"enable" usage:"enable NFS support"`
	Port   int  `json:"port"   usage:"which port the NFS server should be bound to" default:"2049"`
	//TLS  livetls.Config `json:"tls"`
}

func ListenAndServeNFS(ctx context.Context, cfg NFSServerConfig, fs afero.Fs) error {
	address := fmt.Sprintf("0.0.0.0:%d", cfg.Port)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	lifecycle.Defer(func(ctx context.Context) {
		_ = listener.Close()
	})

	handler := helpers.NewNullAuthHandler(afs.Billy(fs))
	handler = helpers.NewCachingHandler(handler, 1024)
	// it would be nice if I could use it without this caching handler

	go func() {
		err = nfs.Serve(listener, handler)
		if err != nil {
			zaputil.Extract(ctx).Error("failed to serve tcp", zap.Error(err))
		}
	}()

	return nil
}
