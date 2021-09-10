// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

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
