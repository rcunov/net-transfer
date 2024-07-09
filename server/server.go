package main

import (
	"crypto/rand"
	"crypto/tls"
	"log"
	"net"
	"rcunov/net-transfer/utils"
)

// Declare globally so main() and tests always use the same values
const (
	certFile = "server.pem"
	keyFile  = "server.key"
	port     = "6600"
)

// StartServer starts listening on the assigned port using TLS with the provided certificate and private key.
func StartServer(port string, certFile string, keyFile string) (listener net.Listener, err error) {
	cert, err := utils.LoadCert(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	config := tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.RequireAnyClientCert}
	config.Rand = rand.Reader // This is default behavior but want to make sure this stays the same

	listener, err = tls.Listen("tcp", ":"+port, &config)
	if err != nil {
		return nil, err
	}

	return listener, nil
}

// HandleConnection sends a simple greeting to the provided connection.
func HandleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("client connected over TLS from %v", conn.RemoteAddr())

	rw := utils.CreateReadWriter(conn)

	// Read message from client
	clientMessage, err := rw.ReadString('\n')
	if err != nil {
		log.Println("client disconnected")
		return
	}

	// Send response to client
	serverResponse := ""
	if clientMessage == "upload\n" {
		log.Print("client wants to upload something")
		serverResponse = "okay"
	} else {
		log.Print("invalid request. client said: ", clientMessage)
		serverResponse = "bad"
	}
	serverResponse = serverResponse + "\n"
	log.Print("responding with: ", serverResponse)
	rw.WriteString(serverResponse)
	rw.Flush()
}

func main() {
	server, err := StartServer(port, certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	log.Println("server listening on port", port)

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go HandleConnection(conn)
	}
}
