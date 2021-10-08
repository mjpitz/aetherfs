// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package auth

import "context"

// RequireAuthentication provides a HandleFunc who ensures the authenticated user is allowed to issue the request.
// Currently, this function only ensures that the user is authenticated and does no further inspection.
func RequireAuthentication() HandleFunc {
	return func(ctx context.Context) (context.Context, error) {
		userInfo := ExtractUserInfo(ctx)
		if userInfo == nil {
			return nil, errUnauthorized
		}
		return ctx, nil
	}
}
