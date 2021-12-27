// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package commands

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"

	"github.com/mjpitz/aetherfs/internal/components"
	"github.com/mjpitz/aetherfs/internal/storage/local"
	"github.com/mjpitz/myago/auth"
	oidcauth "github.com/mjpitz/myago/auth/oidc"
	"github.com/mjpitz/myago/browser"
	"github.com/mjpitz/myago/flagset"
	"github.com/mjpitz/myago/zaputil"
)

// AuthConfig encapsulates all the configuration required to log in to an AetherFS instance.
type AuthConfig struct {
	components.GRPCClientConfig
}

// Auth returns a cli.Command that can be added to an existing application.
func Auth() *cli.Command {
	cfg := &AuthConfig{
		GRPCClientConfig: components.GRPCClientConfig{
			OIDC: oidcauth.Config{
				RedirectURL: "http://localhost:23843/callback",
			},
		},
	}

	return &cli.Command{
		Name:      "auth",
		Usage:     "Manage authentication to AetherFS servers",
		UsageText: "aetherfs auth <command>",
		Subcommands: []*cli.Command{
			{
				Name:      "add",
				Usage:     "Adds authentication to an AetherFS server",
				UsageText: "aetherfs auth add [options] <server>",
				Flags:     flagset.Extract(cfg),
				Action: func(ctx *cli.Context) error {
					db := local.Extract(ctx.Context)
					credentials := db.Credentials()
					tokens := db.Tokens()

					server := ctx.Args().Get(0)
					if len(server) == 0 {
						return fmt.Errorf("server name not provided")
					}

					if len(cfg.Target) == 0 {
						cfg.Target = server
					}

					err := credentials.Put(ctx.Context, server, cfg.GRPCClientConfig)

					if err != nil {
						return err
					} else if cfg.GRPCClientConfig.AuthType != "oidc" {
						return nil
					}

					// oidc only here on out

					uri, err := url.Parse(cfg.GRPCClientConfig.OIDC.RedirectURL)
					if err != nil {
						return err
					}

					svr := &http.Server{
						Addr: uri.Host,
					}

					if len(cfg.GRPCClientConfig.OIDC.Scopes.Value()) == 0 {
						cfg.GRPCClientConfig.OIDC.Scopes = cli.NewStringSlice("openid", "profile", "email")
					}

					cctx, cancel := context.WithCancel(ctx.Context)

					svr.Handler = oidcauth.ServeMux(cfg.GRPCClientConfig.OIDC, func(token *oauth2.Token) {
						defer cancel()

						err = tokens.Put(ctx.Context, cfg.Target, token)
						if err != nil {
							zaputil.Extract(ctx.Context).Error("failed to store token", zap.Error(err))
						}
					})

					group := &errgroup.Group{}

					group.Go(func() error {
						time.Sleep(time.Second)
						return browser.Open(ctx.Context, uri.Scheme+"://"+uri.Host+"/login")
					})

					group.Go(svr.ListenAndServe)

					<-cctx.Done()
					err = svr.Shutdown(ctx.Context)
					_ = group.Wait()
					return nil
				},
				HideHelpCommand: true,
			},
			{
				Name:      "remove",
				Usage:     "Removes authentication to an AetherFS server",
				UsageText: "aetherfs auth remove <server>",
				Action: func(ctx *cli.Context) error {
					credentials := local.Extract(ctx.Context).Credentials()

					server := ctx.Args().Get(0)
					if len(server) == 0 {
						return fmt.Errorf("server name not provided")
					}

					return credentials.Delete(ctx.Context, server)
				},
				HideHelpCommand: true,
			},
			{
				Name:      "show",
				Usage:     "Shows non-sensitive authentication information for an AetherFS server",
				UsageText: "aetherfs auth show <server>",
				Action: func(ctx *cli.Context) error {
					db := local.Extract(ctx.Context)
					credentials := db.Credentials()
					tokens := db.Tokens()

					server := ctx.Args().Get(0)
					if len(server) == 0 {
						return fmt.Errorf("server name not provided")
					}

					clientConfig := components.GRPCClientConfig{}
					err := credentials.Get(ctx.Context, server, &clientConfig)
					if err != nil {
						return err
					}

					t, err := template.New("details").Parse(details)
					if err != nil {
						return err
					}

					userInfo := auth.UserInfo{
						Subject: clientConfig.Basic.Username,
						Profile: clientConfig.Basic.Username,
					}

					if clientConfig.AuthType == "oidc" {
						token := &oauth2.Token{}
						err = tokens.Get(ctx.Context, clientConfig.Target, token)
						if err != nil {
							return err
						}

						provider, err := clientConfig.OIDC.Issuer.Provider(ctx.Context)
						if err != nil {
							return err
						}

						info, err := provider.UserInfo(ctx.Context, oauth2.StaticTokenSource(token))
						if err != nil {
							return err
						}

						userInfo = auth.UserInfo{}
						err = info.Claims(&userInfo)
						if err != nil {
							return err
						}
					}

					return t.Execute(ctx.App.Writer, &data{
						Server:           server,
						GRPCClientConfig: clientConfig,
						UserInfo:         userInfo,
					})
				},
				HideHelpCommand: true,
			},
		},
		HideHelpCommand: true,
	}
}

type data struct {
	Server string
	components.GRPCClientConfig
	auth.UserInfo
}

const details = `
SERVER: {{ .Server }}
TARGET: {{ .GRPCClientConfig.Target }}
{{ if .GRPCClientConfig.TLS.Enable }}
TLS
===
{{- if .GRPCClientConfig.TLS.CertPath }}
CERT PATH: {{ .GRPCClientConfig.TLS.CertPath }}
CA FILE:   {{ .GRPCClientConfig.TLS.CAFile }}
CERT FILE: {{ .GRPCClientConfig.TLS.CertFile }}
KEY FILE:  {{ .GRPCClientConfig.TLS.KeyFile }}
{{- else }}
CERT PATH: <system>
{{- end }}
{{- end }}
{{ if .UserInfo.Subject }}
USER
====
PROFILE:   {{ .UserInfo.Profile }}
{{- if .UserInfo.Email }}
EMAIL:     {{ .UserInfo.Email }}
VERIFIED:  {{ .UserInfo.EmailVerified }}
{{- end }}
AUTH_TYPE: {{ .GRPCClientConfig.Config.AuthType }}
{{- if eq .GRPCClientConfig.Config.AuthType "oidc" }}
ISSUER:    {{ .GRPCClientConfig.OIDC.Issuer.ServerURL }}
{{- if .GRPCClientConfig.OIDC.Issuer.CertificateAuthority }}
CA:        {{ .GRPCClientConfig.OIDC.Issuer.CertificateAuthority }}
{{- end }}
CLIENT_ID: {{ .GRPCClientConfig.OIDC.ClientID }}
{{- end }}
{{- with .UserInfo.Groups }}
GROUPS:
{{- range $group := . }}
- {{ $group }}
{{- end }}
{{- end }}
{{ end }}
`
