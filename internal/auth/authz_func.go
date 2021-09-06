package auth

import "context"

// AuthorizationFunc provides a HandleFunc who ensures the authenticated user is allowed to issue the request.
// Currently, this function only ensures that the user is authenticated and does no further inspection.
func AuthorizationFunc() HandleFunc {
	return func(ctx context.Context) (context.Context, error) {
		userInfo := ExtractUserInfo(ctx)
		if userInfo == nil {
			return nil, errUnauthorized
		}
		return ctx, nil
	}
}