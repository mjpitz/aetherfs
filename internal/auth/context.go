package auth

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
)

type contextKey string

const userInfoContextKey = contextKey("user_info")

// ExtractUserInfo will extract the oidc.UserInfo from the request. This function assumes the AuthenticatorOIDC has
// run. If it hasn't then the UserInfo
func ExtractUserInfo(ctx context.Context) *oidc.UserInfo {
	v := ctx.Value(userInfoContextKey)
	if v == nil {
		return nil
	}

	userInfo, ok := v.(*oidc.UserInfo)
	if !ok {
		return nil
	}

	return userInfo
}
