package utils

import (
	"crypto/tls"
)

// LoadCert uses tls.LoadX509KeyPair() to parse a public/private key pair from a pair of files. If an error is found, LoadCert will return an empty certificate and the error.
func LoadCert(certFile string, keyFile string) (cert tls.Certificate, err error) {
	cert, err = tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		return tls.Certificate{}, err
	}

	return cert, nil
}