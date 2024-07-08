package main

import (
	"crypto/tls"
	"io"
	"log"
	"rcunov/net-transfer/utils"
)

func main() {
	cert, err := utils.LoadCert("client.pem", "client.key")
	if err != nil {
		log.Fatal(err)
	}

	// server connection info
	hostname := "localhost" // TODO: Change me
	port := "6600"
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

	log.Printf("Connecting to server at %s\n", hostname)
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