// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package components

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

// TranslateHeadersToMetadata provides a gin handler that will copy Headers from an HTTP request to gRPC headers.
func TranslateHeadersToMetadata() gin.HandlerFunc {
	return func(ginctx *gin.Context) {
		// preprocess headers into grpc metadata
		md := metadata.New(nil)
		for k, vv := range ginctx.Request.Header {
			md.Set(k, vv...)
		}

		ctx := metadata.NewIncomingContext(ginctx.Request.Context(), md)
		ginctx.Request = ginctx.Request.WithContext(ctx)
	}
}
