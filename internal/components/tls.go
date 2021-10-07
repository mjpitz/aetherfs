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
