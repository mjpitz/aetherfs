// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package lifecycle

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var once = sync.Once{}

// Setup initializes a shutdown hook that cancels the underlying context.
func Setup(ctx context.Context) context.Context {
	once.Do(func() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)

		halt := make(chan os.Signal, 1)
		signal.Notify(halt, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-halt
			signal.Stop(halt)

			cancel()
		}()
	})

	return ctx
}
