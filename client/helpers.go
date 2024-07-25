package main

import (
	"bufio"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

// ConnectToServer initiates a TLS connection to the server at the provided hostname and port.
func ConnectToServer(tlsCert tls.Certificate, hostname string, port string) (conn net.Conn, err error) {
	fmt.Printf("Connecting to server at %v\n", hostname)
	config := tls.Config{Certificates: []tls.Certificate{tlsCert}, InsecureSkipVerify: true}
	conn, err = tls.Dial("tcp", hostname+":"+port, &config)
	if err != nil {
		return nil, err
	}

	localPort := conn.LocalAddr().(*net.TCPAddr).Port
	fmt.Printf("Connection established from localhost:%v --> %v:%v\n", localPort, hostname, port)

	return conn, err
}

func DisplayCurrentDirectoryFiles() {
	fmt.Println("Files in current directory:")
	files, err := os.ReadDir(".")
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}
	for _, file := range files {
		if !file.IsDir() {
			fmt.Println(file.Name())
		}
	}
}

func GetNumberOfFiles(rw *bufio.ReadWriter) (int, error) {
	numFilesStr, err := rw.ReadString('\n')
	if err != nil {
		return 0, err
	}
	numFiles, err := strconv.Atoi(numFilesStr[:len(numFilesStr)-1])
	if err != nil {
		return 0, err
	}
	return numFiles, nil
}

func GetListOfFiles(rw *bufio.ReadWriter, numFiles int) ([]string, error) {
	files := make([]string, 0, numFiles)
	for i := 0; i < numFiles; i++ {
		fileName, err := rw.ReadString('\n')
		if err != nil {
			return nil, err
		}
		files = append(files, fileName[:len(fileName)-1])
	}
	return files, nil
}

func DisplayFiles(files []string) {
	fmt.Println("Available files:")
	for i, file := range files {
		fmt.Printf("%d: %s\n", i+1, file)
	}
}

func GetUserSelection(options int) (int, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the number of the file to download (or press CTRL-C to cancel): ")
	for {
		selectionStr, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Println()
			fmt.Println("Goodbye!")
			os.Exit(0)
		}
		if err != nil {
			fmt.Println("Error reading input. Please try again.")
			continue
		}
		if selectionStr == "\n" {
			// I don't know why, but when you open the download menu for the first time, the program
			// reads in a newline character from stdin. Ignoring newline characters from stdin will fix
			// this, but causes an issue where spamming the Enter key will push the dialog up the screen
			// since it only prints once, and if we try to catch newline characters with a re-print of
			// the selection menu, it will double-print the menu when you first open the program. I've
			// chosen to stick with ignoring them and hoping a user doesn't spam the Enter key.
			continue
		}
		selectionStr = strings.TrimSpace(selectionStr) // Remove newline character from the end
		selection, err := strconv.Atoi(selectionStr)
		if err != nil || selection <= 0 || selection > options {
			fmt.Printf("Invalid selection.\nPlease enter a number between 1 and %d or CTRL-C to cancel: ", options)
			continue
		}
		return selection, nil
	}
}

func SendSelection(rw *bufio.ReadWriter, selection int) error {
	_, err := rw.WriteString(strconv.Itoa(selection) + "\n")
	if err != nil {
		return err
	}
	return rw.Flush()
}

func ReceiveFileSizeAndHash(rw *bufio.ReadWriter) (int64, string, error) {
	fileSizeStr, err := rw.ReadString('\n')
	if err != nil {
		return 0, "", err
	}
	fileSize, err := strconv.ParseInt(fileSizeStr[:len(fileSizeStr)-1], 10, 64)
	if err != nil {
		return 0, "", err
	}

	fileHash, err := rw.ReadString('\n')
	if err != nil {
		return 0, "", err
	}
	fileHash = fileHash[:len(fileHash)-1]

	return fileSize, fileHash, nil
}

func ReceiveFile(rw *bufio.ReadWriter, fileName string, fileSize int64, expectedHash string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	err = file.Truncate(fileSize)
	if err != nil {
		return err
	}

	_, err = io.CopyN(file, rw.Reader, fileSize)
	if err != nil {
		return err
	}

	file.Seek(0, 0)
	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return err
	}
	calculatedHash := hex.EncodeToString(hash.Sum(nil))
	if calculatedHash != expectedHash {
		return fmt.Errorf("file hash mismatch: expected %s, got %s", expectedHash, calculatedHash)
	}
	return nil
}

func SendFile(rw *bufio.ReadWriter, fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(rw.Writer, file)
	if err != nil {
		return err
	}
	return rw.Flush()
}
