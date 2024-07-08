package main

import (
	"crypto/rand"
	"crypto/tls"
	"log"
	"net"
	"rcunov/net-transfer/utils"
)

// HandleConnection sends a simple greeting to the provided connection.
func HandleConnection(c net.Conn) {
 	log.Printf("Client connected over TLS from %v", c.RemoteAddr())
	c.Write([]byte("hello :)"))
	c.Close()
	log.Printf("Greetings sent to %v. Connection closed", c.RemoteAddr())
}

func main() {
	cert, err := utils.LoadCert("server.pem", "server.key")
	if err != nil {
		log.Fatal(err)
	}
	
 	config := tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.RequireAnyClientCert}
 	config.Rand = rand.Reader // this is default behavior, but want to make sure this stays the same

	port := "6600"
 	ln, err := tls.Listen("tcp", ":"+port, &config)
 	if err != nil {
 		log.Fatal(err)
 	}

 	log.Println("Server listening on port", port)

 	for {
 		conn, err := ln.Accept()
 		if err != nil {
 			log.Println(err)
 			continue
 		}
 		go HandleConnection(conn)
 	}
}