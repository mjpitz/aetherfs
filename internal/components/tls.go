// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package components

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
)

// TLSConfig defines common configuration that can be used to LoadCertificates for encrypted communication.
type TLSConfig struct {
	Enable               bool   `json:"enable,omitempty"                usage:"enable TLS communication"`
	CertificateAuthority string `json:"certificate_authority,omitempty" usage:"specify the certificate authority"`
	Certificate          string `json:"certificate,omitempty"           usage:"specify the certificate to use for communication"`
	PrivateKey           string `json:"private_key,omitempty"           usage:"specify the private key to use for communication"`
}

// LoadCertificates uses the provided TLSConfig to load the various certificates and return a proper tls.Config. This
// method works a handful of ways. The tls.Config produced by this method can be used by either the client or the
// server.
//
// First, you can set Enable to return a tls.Config that uses the systems certificate authority. This is most
// useful when communicating with a public facing service who's certificate is backed by something like LetsEncrypt.
//
// Second, you can just customize the CertificateAuthority that's used.
//
// Finally, you can specify the CertificateAuthority, Certificate, and PrivateKey the system should use. This approach
// is most commonly used when deployed with mTLS. Servers will need to explicitly set
// `cfg.ClientAuth = tls.RequireAndVerifyClientCert` to enable mTLS.
//
func LoadCertificates(cfg TLSConfig) (*tls.Config, error) {
	if !cfg.Enable {
		return nil, nil
	}

	var caCertPool *x509.CertPool
	var certificates []tls.Certificate

	if cfg.CertificateAuthority != "" {
		caCert, err := ioutil.ReadFile(cfg.CertificateAuthority)
		if err != nil {
			return nil, err
		}

		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
	}

	if cfg.Certificate != "" && cfg.PrivateKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.Certificate, cfg.PrivateKey)
		if err != nil {
			return nil, err
		}

		certificates = append(certificates, cert)
	}

	return &tls.Config{
		RootCAs:      caCertPool, // used by clients to verify servers
		ClientCAs:    caCertPool, // used by servers to verify clients
		Certificates: certificates,
	}, nil
}
