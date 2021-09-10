// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

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