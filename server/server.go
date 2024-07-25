// server.go
package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"rcunov/net-transfer/utils"
	"strconv"
)

var netErr *net.OpError // Used to catch connection termination error

const (
	port = "6600"
)

// StartServer starts listening on the assigned port using TLS with the provided certificate and private key.
func StartServer(port string) (listener net.Listener, err error) {
	cert, err := utils.GenerateCert()
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

func main() {
	listener, err := StartServer(port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server is listening on port " + port)

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
			err = HandleFileDownload(rw)
			if err != nil {
				fmt.Printf("Error handling client %v file download: %v\n", remoteAddr, err)
				return
			}
		case "2":
			err = HandleFileUpload(rw)
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

func HandleFileDownload(rw *bufio.ReadWriter) error {
	files := []string{"file1.txt", "file2.txt", "file3.txt"}

	err := SendNumberOfFiles(rw, len(files))
	if errors.As(err, &netErr) { // Abrubt connection termination
		return fmt.Errorf("client has disconnected")
	}
	if err != nil {
		return err
	}

	err = SendListOfFiles(rw, files)
	if errors.As(err, &netErr) {
		return fmt.Errorf("client has disconnected")
	}
	if err != nil {
		return err
	}

	selection, err := GetClientSelection(rw)
	if errors.As(err, &netErr) {
		return fmt.Errorf("client has disconnected")
	}
	if err != nil {
		return err
	}

	if selection < 1 || selection > len(files) {
		return fmt.Errorf("invalid selection")
	}

	fileName := files[selection-1]
	fileSize, fileHash, err := GetFileSizeAndHash(fileName)
	if errors.As(err, &netErr) {
		return fmt.Errorf("client has disconnected")
	}
	if err != nil {
		return err
	}

	err = SendFileSizeAndHash(rw, fileSize, fileHash)
	if errors.As(err, &netErr) {
		return fmt.Errorf("client has disconnected")
	}
	if err != nil {
		return err
	}

	err = SendFile(rw, fileName)
	if errors.As(err, &netErr) {
		return fmt.Errorf("client has disconnected")
	}
	if err != nil {
		return err
	}

	fmt.Println("File sent successfully")
	return nil
}

func HandleFileUpload(rw *bufio.ReadWriter) error {
	fileName, err := rw.ReadString('\n')

	if errors.As(err, &netErr) { // Abrubt connection termination
		return fmt.Errorf("client has disconnected")
	}
	if err != nil {
		return err
	}

	fileName = fileName[:len(fileName)-1]
	fileSizeStr, err := rw.ReadString('\n')
	if errors.As(err, &netErr) {
		return fmt.Errorf("client has disconnected")
	}
	if err != nil {
		return err
	}
	fileSize, err := strconv.ParseInt(fileSizeStr[:len(fileSizeStr)-1], 10, 64)
	if err != nil {
		return err
	}

	// TODO: Do something more elegant than read for exact input, maybe like a (y/N) where
	// anything starting with "y" accepts and anything starting with "n" or empty declines
	fmt.Printf("Client wants to upload %s (%v). Accept? (yes/no): ", fileName, BytesPrettyPrint(fileSize))
	var approval string
	fmt.Scan(&approval)

	if approval != "yes" {
		_, err := rw.WriteString("no\n")
		if errors.As(err, &netErr) {
			return fmt.Errorf("client has disconnected")
		}
		if err != nil {
			return err
		}
		rw.Flush()
		return nil
	}

	_, err = rw.WriteString("yes\n")
	if errors.As(err, &netErr) {
		return fmt.Errorf("client has disconnected")
	}
	if err != nil {
		return err
	}
	rw.Flush()

	fileHash, err := rw.ReadString('\n')
	if errors.As(err, &netErr) {
		return fmt.Errorf("client has disconnected")
	}
	if err != nil {
		return err
	}
	fileHash = fileHash[:len(fileHash)-1]

	err = ReceiveFile(rw, fileName, fileSize, fileHash)
	if errors.As(err, &netErr) {
		return fmt.Errorf("client has disconnected")
	}
	if err != nil {
		return err
	}

	fmt.Println("File received and verified successfully")
	return nil
}

func SendNumberOfFiles(rw *bufio.ReadWriter, numFiles int) error {
	_, err := rw.WriteString(strconv.Itoa(numFiles) + "\n")
	if err != nil {
		return err
	}
	return rw.Flush()
}

func SendListOfFiles(rw *bufio.ReadWriter, files []string) error {
	for _, file := range files {
		_, err := rw.WriteString(file + "\n")
		if err != nil {
			return err
		}
	}
	return rw.Flush()
}

func GetClientSelection(rw *bufio.ReadWriter) (int, error) {
	selectionStr, err := rw.ReadString('\n')
	if err != nil {
		return 0, err
	}
	selection, err := strconv.Atoi(selectionStr[:len(selectionStr)-1])
	if err != nil {
		return 0, err
	}
	return selection, nil
}

func GetFileSizeAndHash(fileName string) (int64, string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return 0, "", err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return 0, "", err
	}
	fileSize := fileInfo.Size()

	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return 0, "", err
	}
	fileHash := hex.EncodeToString(hash.Sum(nil))

	return fileSize, fileHash, nil
}

func BytesPrettyPrint(bytes int64) string { // Shamelessly stolen
	const base = 1000
	if bytes < base {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(base), 0
	for n := bytes / base; n >= base; n /= base {
		div *= base
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "kMGTPE"[exp])
}

func SendFileSizeAndHash(rw *bufio.ReadWriter, fileSize int64, fileHash string) error {
	_, err := rw.WriteString(strconv.FormatInt(fileSize, 10) + "\n")
	if err != nil {
		return err
	}
	rw.Flush()

	_, err = rw.WriteString(fileHash + "\n")
	if err != nil {
		return err
	}
	return rw.Flush()
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
