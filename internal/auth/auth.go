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
