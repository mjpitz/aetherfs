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
	"os"
	"path/filepath"
)

// TLSConfig defines common configuration that can be used to LoadCertificates for encrypted communication.
type TLSConfig struct {
	CertPath string `json:"cert_path" usage:"where to locate certificates for communication"`
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
	if cfg.CertPath == "" {
		return nil, nil
	}

	caPath := filepath.Join(cfg.CertPath, "ca.pem")
	certPath := filepath.Join(cfg.CertPath, "cert.pem")
	keyPath := filepath.Join(cfg.CertPath, "key.pem")

	_, caPathErr := os.Stat(caPath)
	_, certPathErr := os.Stat(certPath)
	_, keyPathErr := os.Stat(keyPath)

	var caCertPool *x509.CertPool
	var certificates []tls.Certificate

	if caPathErr == nil {
		caCert, err := ioutil.ReadFile(caPath)
		if err != nil {
			return nil, err
		}

		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
	}

	if certPathErr == nil && keyPathErr == nil {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
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
