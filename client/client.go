// client.go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"rcunov/net-transfer/utils"
	"strconv"
)

// Server connection info
var (
	portFlag = flag.String("port", "6600", "Set the port to listen on (defaults to 6600)")
	hostname = flag.String("hostname", "localhost", "Set the hostname to connect to (defaults to localhost)")
)

func main() {
	flag.Parse()
	if !utils.IsValidPort(*portFlag) {
		fmt.Printf("invalid port specified: %v. should be 1025-65535", *portFlag)
		os.Exit(1)
	}

	tlsCert, err := utils.GenerateCert()
	if err != nil {
		fmt.Printf("Error: cannot generate the certificate: %v\n", err.Error())
		os.Exit(1)
	}

	conn, err := ConnectToServer(tlsCert, *hostname, *portFlag)
	if err != nil {
		fmt.Printf("Error: cannot connect to server: %v\n", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	for {
		fmt.Println("1: Download a file")
		fmt.Println("2: Upload a file")
		fmt.Println("CTRL-C: Quit")
		fmt.Print("Enter your selection: ")

		selection, err := GetUserSelection(2)
		if err != nil {
			fmt.Println("Error getting selection:", err)
			return
		}

		_, err = rw.WriteString(strconv.Itoa(selection) + "\n")
		if err != nil {
			fmt.Println("Error sending selection:", err)
			return
		}
		rw.Flush()

		switch selection {
		case 1:
			err = HandleFileDownload(rw)
			if err == io.EOF {
				fmt.Println()
				fmt.Println("Goodbye!")
				os.Exit(0)
			}
			if err != nil {
				fmt.Println("Error handling file download:", err)
				return
			}
			fmt.Println()
			fmt.Println("Goodbye!")
			os.Exit(0)
		case 2:
			DisplayCurrentDirectoryFiles()
			err = HandleFileUpload(rw)
			if err == io.EOF {
				fmt.Println()
				fmt.Println("Goodbye!")
				os.Exit(0)
			}
			if err != nil {
				fmt.Println("Error handling file upload:", err)
				return
			}
			fmt.Println()
			fmt.Println("Goodbye!")
			os.Exit(0)
		}
	}
}
