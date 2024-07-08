package main

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"log"
	"net"
	"rcunov/net-transfer/utils"
)

// HandleConnection sends a simple greeting to the provided connection.
func HandleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Client connected over TLS from %v", conn.RemoteAddr())

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	serverMessage := ""

	// Read message from client
	clientMessage, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Client disconnected.")
		return
	}

	// Send response to client
	if clientMessage == "upload\n" {
		log.Print("Client wants to upload something")
		serverMessage = "okay"
	} else {
		log.Print("invalid request. client said: ", clientMessage)
		serverMessage = "bad"
	}
	serverMessage = serverMessage + "\n"
	log.Print("Responding with: ", serverMessage)
	writer.WriteString(serverMessage)
	writer.Flush()
}

// Declare key pair locations globally so main() and tests use the same paths
const (
	certFile = "server.pem"
	keyFile  = "server.key"
)

const port = "6600"

func main() {
	cert, err := utils.LoadCert(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}

	config := tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.RequireAnyClientCert}
	config.Rand = rand.Reader // This is default behavior but want to make sure this stays the same

	listen, err := tls.Listen("tcp", ":"+port, &config)
	if err != nil {
		log.Fatal(err)
	}
	defer listen.Close()

	log.Println("Server listening on port", port)

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go HandleConnection(conn)
	}
}
