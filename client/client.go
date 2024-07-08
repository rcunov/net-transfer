package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"time"
)

// Server connection info
var (
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
		log.Fatal("could not generate random serial for cert. error:", err)
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

func main() {
	tlsCert, err := GenerateCert()
	if err != nil {
		log.Fatal("cannot generate the certificate.", err.Error())
	}

	log.Printf("Connecting to server at %s\n", hostname)
	config := tls.Config{Certificates: []tls.Certificate{tlsCert}, InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", hostname+":"+port, &config)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	log.Println("Established connection to server")

	message, err := io.ReadAll(conn)
	if err != nil {
		log.Println("error reading content from server:", err.Error())
		return
	}
	log.Printf("Message received: %s\n", message)
}
