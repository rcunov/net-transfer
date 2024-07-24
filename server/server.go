// server.go
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server is listening on port 8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			return
		}
		fmt.Println("Client connected")

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	files := []string{"file1.txt", "file2.txt", "file3.txt"}

	err := SendNumberOfFiles(rw, len(files))
	if err != nil {
		fmt.Println("Error sending number of files:", err)
		return
	}

	err = SendListOfFiles(rw, files)
	if err != nil {
		fmt.Println("Error sending list of files:", err)
		return
	}

	choice, err := GetClientChoice(rw)
	if err != nil {
		fmt.Println("Error reading choice:", err)
		return
	}

	if choice < 1 || choice > len(files) {
		fmt.Println("Invalid choice")
		return
	}

	fileName := files[choice-1]
	fileSize, fileHash, err := GetFileSizeAndHash(fileName)
	if err != nil {
		fmt.Println("Error getting file size and hash:", err)
		return
	}

	err = SendFileSizeAndHash(rw, fileSize, fileHash)
	if err != nil {
		fmt.Println("Error sending file size and hash:", err)
		return
	}

	err = SendFile(rw, fileName)
	if err != nil {
		fmt.Println("Error sending file:", err)
		return
	}
	fmt.Println("File sent successfully")
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

func GetClientChoice(rw *bufio.ReadWriter) (int, error) {
	choiceStr, err := rw.ReadString('\n')
	if err != nil {
		return 0, err
	}
	choice, err := strconv.Atoi(choiceStr[:len(choiceStr)-1])
	if err != nil {
		return 0, err
	}
	return choice, nil
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
