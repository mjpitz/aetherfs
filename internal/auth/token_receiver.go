// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package auth

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// TokenCallback is invoked by the TokenReceiver endpoint when we've successfully received and validated the
// authenticated user session.
type TokenCallback func(token *oauth2.Token)

// TokenReceiver is some rough code that should allow a command line tool to receive a token and invoke the provided
// callback function when a successful exchange is performed.
func TokenReceiver(cfg OIDCConfig, callback TokenCallback) gin.HandlerFunc {
	mu := &sync.Mutex{}

	var provider *oidc.Provider
	var verifier *oidc.IDTokenVerifier

	obtainProviderAndVerifier := func(ctx context.Context) (*oidc.Provider, *oidc.IDTokenVerifier, error) {
		mu.Lock()
		defer mu.Unlock()

		if provider == nil {
			var err error
			provider, err = oidc.NewProvider(ctx, cfg.Issuer.ServerURL)
			if err != nil {
				return nil, nil, err
			}

			verifier = provider.Verifier(&oidc.Config{
				ClientID: cfg.ClientID,
			})
		}

		return provider, verifier, nil
	}

	oauth2Config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return func(ctx *gin.Context) {
		logger := ctxzap.Extract(ctx)

		_, verifier, err := obtainProviderAndVerifier(ctx)
		if err != nil {
			logger.Error("error establishing connection with provider", zap.Error(err))
			_ = ctx.AbortWithError(http.StatusInternalServerError, nil)
			return
		}

		oauth2Token, err := oauth2Config.Exchange(ctx, ctx.Query("code"))
		if err != nil {
			logger.Error("failed to exchange code for auth info", zap.Error(err))
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !oauth2Token.Valid() || !ok {
			_ = ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("id_token missing or invalid"))
			return
		}

		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		userInfo := &oidc.UserInfo{}
		err = idToken.Claims(userInfo)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		ctx.String(http.StatusOK, "You have successfully logged in. You may now close this tab in your browser.")
		go callback(oauth2Token)
	}
}
