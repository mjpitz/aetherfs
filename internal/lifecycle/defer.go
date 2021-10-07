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
	"sync"
)

var lock = sync.Mutex{}
var funcs = make([]func(ctx context.Context), 0)

// Defer will enqueue a function that will be invoked by Resolve.
func Defer(fn func(ctx context.Context)) {
	lock.Lock()
	defer lock.Unlock()

	funcs = append(funcs, fn)
}

// Resolve will process all functions that have been enqueued by Defer up until this point.
func Resolve(ctx context.Context) {
	lock.Lock()
	defer lock.Unlock()

	for i := len(funcs); i > 0; i-- {
		funcs[i-1](ctx)
	}

	funcs = funcs[len(funcs):]
}
