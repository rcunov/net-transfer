package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
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
func ConnectToServer(tlsCert tls.Certificate, hostname string, port string) (conn *tls.Conn, err error) {
	log.Printf("Connecting to server at %s\n", hostname)
	config := tls.Config{Certificates: []tls.Certificate{tlsCert}, InsecureSkipVerify: true}
	conn, err = tls.Dial("tcp", hostname+":"+port, &config)
	if err != nil {
		return nil, err
	}
	log.Println("Established connection to server")

	return conn, err
}

func main() {
	tlsCert, err := GenerateCert()
	if err != nil {
		log.Fatal("cannot generate the certificate.", err.Error())
	}

	conn, err := ConnectToServer(tlsCert, hostname, port)
	if err != nil {
		log.Fatal("cannot connect to server. ", err)
	}

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Send behavior selection
	msg := "upload" + "\n"
	log.Print("Sending message to server: ", msg)
	writer.WriteString(msg)
	writer.Flush()

	// Read response from server
	serverMessage, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Server disconnected.")
		return
	}
	log.Print("Server responded: ", serverMessage)
}
