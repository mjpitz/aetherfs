// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

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
