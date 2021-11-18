// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

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

type Config struct {
	AuthType string `json:"auth_type" usage:"configure the user authentication type to use"`
}
