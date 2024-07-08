package utils

import (
	"bufio"
	"crypto/tls"
	"net"
)

// LoadCert uses tls.LoadX509KeyPair() to parse a public/private key pair from a pair of files. If an error is found, LoadCert will return an empty certificate and the error.
func LoadCert(certFile string, keyFile string) (cert tls.Certificate, err error) {
	cert, err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return tls.Certificate{}, err
	}

	return cert, nil
}

// CreateReadWriter allows us to read and write to a connection.
func CreateReadWriter(conn net.Conn) (readwriter *bufio.ReadWriter) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	readwriter = bufio.NewReadWriter(reader, writer)
	return readwriter
}