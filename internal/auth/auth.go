// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package auth

import "context"

// Composite returns a handler that iterates all HandleFuncs.
func Composite(fns ...HandleFunc) HandleFunc {
	return func(ctx context.Context) (context.Context, error) {
		var err error
		for _, fn := range fns {
			ctx, err = fn(ctx)
			if err != nil {
				return nil, err
			}
		}
		return ctx, nil
	}
}
