// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package local

import (
	"context"

	"github.com/mjpitz/myago"
)

const contextKey = myago.ContextKey("local.db")

// Extract returns the local db instance on the context (if present).
func Extract(ctx context.Context) *DB {
	v := ctx.Value(contextKey)
	if v == nil {
		return nil
	}

	return v.(*DB)
}
