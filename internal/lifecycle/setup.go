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
