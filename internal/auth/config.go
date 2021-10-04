// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package auth

type OIDCIssuer struct {
	ServerURL            string `json:"server_url"            usage:"the address of the server where user authentication is performed"`
	CertificateAuthority string `json:"certificate_authority" usage:"path pointing to a file containing the certificate authority data for the server"`
}

type OIDCConfig struct {
	Issuer       OIDCIssuer `json:"issuer"`
	ClientID     string     `json:"client_id"     usage:"the client_id associated with this service"`
	ClientSecret string     `json:"client_secret" usage:"the client_secret associated with this service"`
	RedirectURL  string     `json:"redirect_url"  usage:"the redirect_url used by this service to obtain a token"`
}

type OIDCClientConfig struct {
	Issuer OIDCIssuer `json:"issuer"`
}

// Config is predominantly used by the `login` command, but is located here to keep the structure between it and
// ClientConfig similar.
type Config struct {
	AuthType string     `json:"auth_type"      usage:"configure the user authentication type to use"`
	OIDC     OIDCConfig `json:"oidc"`
}

// ClientConfig is used by most processes and handles verifying user authentication. For the most part, we expect our
// system to leverage `dex` to handle federated identity management.
type ClientConfig struct {
	AuthType string           `json:"auth_type"      usage:"configure the user authentication type to use"`
	OIDC     OIDCClientConfig `json:"oidc"`
}
