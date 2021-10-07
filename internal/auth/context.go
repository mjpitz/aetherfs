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

	"github.com/coreos/go-oidc/v3/oidc"
)

type contextKey string

const userInfoContextKey = contextKey("user_info")

// ExtractUserInfo will extract the oidc.UserInfo from the request. This function assumes the OIDCAuthenticator has
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
