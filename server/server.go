package main

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"rcunov/net-transfer/utils"
	"strings"
	"time"
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

func HandleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("client connected over TLS from %v", conn.RemoteAddr())

	rw := utils.CreateReadWriter(conn)
	menu := "--- Menu: ---\n1. HELLO\n2. TIME\n3. EXIT\nEnter your choice: \f"

	for {
		rw.WriteString(menu)
		rw.Flush()
		message, err := rw.ReadString('\n')
		if err != nil {
			log.Println("Connection closed.")
			return
		}

		message = strings.TrimSpace(message)
		log.Printf("Received command: %s\n", message)

		var response string
		switch message {
		case "1":
			response = "Hello, client!\n"
		case "2":
			response = fmt.Sprintf("The current time is: %s\n", time.Now().Format(time.RFC1123))
		case "3":
			response = "Goodbye!\n"
			rw.WriteString(response)
			rw.Flush()
			return
		default:
			response = "Invalid choice. Please select a valid option using the corresponding number.\n"
		}

		rw.WriteString(response)
		rw.Flush()
	}
}

func main() {
	server, err := StartServer(port, certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	log.Print("server listening on port ", port)

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go HandleConnection(conn)
	}
}
