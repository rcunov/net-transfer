package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"rcunov/net-transfer/utils"
	"time"
)

// Server connection info
const (
	hostname = "localhost"
	port     = "6600"
)

// GenerateCert creates a new TLS certificate to encrypt the connection to the server.
// If an error is returned, the certificate will be empty.
func GenerateCert() (cert tls.Certificate, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Generate a pem block with the private key
	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	// Generate random serial number for cert
	serial, err := rand.Int(rand.Reader, big.NewInt(123456))
	if err != nil {
		return tls.Certificate{}, err
	}

	// Generate the cert
	template := x509.Certificate{
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(5, 0, 0),
		// Have to generate a different serial number each execution
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: "net-transfer client",
		},
		BasicConstraintsValid: true,
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Generate a pem block with the cert
	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	// Parse the cert so we can use it
	cert, err = tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return tls.Certificate{}, err
	}

	return cert, err
}

// ConnectToServer initiates a TLS connection to the server at the provided hostname and port.
func ConnectToServer(tlsCert tls.Certificate, hostname string, port string) (conn net.Conn, err error) {
	log.Print("connecting to server at ", hostname)
	config := tls.Config{Certificates: []tls.Certificate{tlsCert}, InsecureSkipVerify: true}
	conn, err = tls.Dial("tcp", hostname+":"+port, &config)
	if err != nil {
		return nil, err
	}
	log.Print("established connection to server")

	return conn, err
}

func main() {
	tlsCert, err := GenerateCert()
	if err != nil {
		log.Fatal("cannot generate the certificate. ", err)
	}

	conn, err := ConnectToServer(tlsCert, hostname, port)
	if err != nil {
		log.Fatal("cannot connect to server. ", err)
	}
	defer conn.Close()

	rw := utils.CreateReadWriter(conn)

	// Send behavior selection
	msg := "upload"
	log.Print("sending message to server: ", msg)
	rw.WriteString(msg + "\n")
	rw.Flush()

	// Read response from server
	serverMessage, err := rw.ReadString('\n')
	if err != nil {
		log.Print("server disconnected")
		return
	}
	log.Print("server responded: ", serverMessage)
}
