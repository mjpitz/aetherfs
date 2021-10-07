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

import (
	"context"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// OIDCAuthenticator returns a HandleFunc who authenticates a user with the provided issuer using an access_token
// attached to the request. If provided, this access_token is exchanged for the authenticated users information. It's
// important to know that this function does not handle authorization and requires an additional HandleFunc to do so.
func OIDCAuthenticator(cfg OIDCIssuer) HandleFunc {
	mu := &sync.Mutex{}
	var provider *oidc.Provider

	obtainProvider := func(ctx context.Context) (*oidc.Provider, error) {
		mu.Lock()
		defer mu.Unlock()

		if provider == nil {
			var err error
			provider, err = oidc.NewProvider(ctx, cfg.ServerURL)
			if err != nil {
				return nil, err
			}
		}

		return provider, nil
	}

	return func(ctx context.Context) (context.Context, error) {
		logger := ctxzap.Extract(ctx)

		provider, err := obtainProvider(ctx)
		if err != nil {
			logger.Error("error establishing connection with provider", zap.Error(err))
			return nil, errInternal
		}

		// ignore error here since some paths might be unauthenticated
		accessToken, _ := grpc_auth.AuthFromMD(ctx, "bearer")
		if accessToken == "" {
			return ctx, nil
		}

		// fetch oidc.UserInfo and put on request
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: accessToken,
		})

		// user is authenticated, but authentication appears to have expired
		// return unauthenticated error here to trigger re-authentication
		userInfo, err := provider.UserInfo(ctx, tokenSource)
		if err != nil {
			logger.Error("error fetching user information given access token", zap.Error(err))
			return nil, errUnauthorized
		}

		// attach user information
		return context.WithValue(ctx, userInfoContextKey, userInfo), nil
	}
}
