package main

import (
	"crypto/tls"
	"io"
	"log"
	"rcunov/net-transfer/utils"
)

// Declare key pair locations globally so main() and tests use the same paths
var (
	certFile = "client.pem"
	keyFile  = "client.key"
)

// Server connection info
var (
	hostname = "localhost"
	port     = "6600"
)

func main() {
	cert, err := utils.LoadCert(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Connecting to server at %s\n", hostname)
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", hostname+":"+port, &config)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	log.Println("Established connection to server")
	
	message, err := io.ReadAll(conn)
	if err != nil {
		log.Println("Error reading content from server:", err.Error())
		return
	}
	log.Printf("Message received: %s\n", message)
}