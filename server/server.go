package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
)

var (
	portFlag = flag.String("port", "6600", "Set the port to listen on (defaults to 6600)")
	netErr   *net.OpError // Used to catch connection termination error
)

func main() {
	flag.Parse()
	listener, err := StartServer(*portFlag)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Printf("Server is listening on port %v\n", *portFlag)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection from %v: %v", conn.RemoteAddr().String(), err)
			return
		}
		fmt.Println("Client connected from %v", conn.RemoteAddr().String())

		go HandleConnection(conn)
	}
}

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	for {
		clientSelection, err := rw.ReadString('\n')
		if errors.As(err, &netErr) {
			fmt.Printf("Client %v has disconnected\n", remoteAddr)
			return
		}
		if err != nil {
			fmt.Printf("Error reading client %v selection: %v\n", remoteAddr, err)
			return
		}
		clientSelection = clientSelection[:len(clientSelection)-1]

		switch clientSelection {
		case "1":
			err = HandleFileDownload(rw, remoteAddr)
			if err != nil {
				fmt.Printf("Error handling client %v file download: %v\n", remoteAddr, err)
				return
			}
		case "2":
			err = HandleFileUpload(rw, remoteAddr)
			if err == io.EOF {
				fmt.Printf("Client %v has disconnected\n", remoteAddr)
				return
			}
			if err != nil {
				fmt.Printf("Error handling client %v file upload: %v", remoteAddr, err)
				return
			}
		default:
			fmt.Printf("Invalid client %v selection\n", remoteAddr)
			return
		}
	}
}
