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

package components

import (
	"context"
	"net/http"
	"time"

	"github.com/Depado/ginprom"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

func GinServer(ctx context.Context) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	ginServer := gin.New()

	ginProm := ginprom.New(
		ginprom.Engine(ginServer),
		ginprom.Path("/metrics"),
	)

	ginServer.Use(
		ginProm.Instrument(),
		ginzap.Ginzap(ctxzap.Extract(ctx), time.RFC3339, true),
		ginzap.RecoveryWithZap(ctxzap.Extract(ctx), true),
	)

	ginServer.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"status": "OK",
		})
	})

	return ginServer
}
