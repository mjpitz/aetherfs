package auth

type OIDCIssuer struct {
	ServerURL            string `json:"server_url,omitempty"            usage:"the address of the server where user authentication is performed"`
	CertificateAuthority string `json:"certificate_authority,omitempty" usage:"path pointing to a file containing the certificate authority data for the server"`
}

type OIDCConfig struct {
	Issuer       OIDCIssuer `json:"issuer,omitempty"`
	ClientID     string     `json:"client_id,omitempty"     usage:"the client_id associated with this service"`
	ClientSecret string     `json:"client_secret,omitempty" usage:"the client_secret associated with this service"`
	RedirectURL  string     `json:"redirect_url,omitempty"  usage:"the redirect_url used by this service to obtain a token"`
}

type Config struct {
	OIDC OIDCConfig `json:"oidc,omitempty"`
}
