// client.go
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
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Connected to server")

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	numFiles, err := GetNumberOfFiles(rw)
	if err != nil {
		fmt.Println("Error getting number of files:", err)
		return
	}

	files, err := GetListOfFiles(rw, numFiles)
	if err != nil {
		fmt.Println("Error getting list of files:", err)
		return
	}

	DisplayFiles(files)

	choice, err := GetUserChoice(len(files))
	if err != nil {
		fmt.Println("Error getting user choice:", err)
		return
	}

	err = SendChoice(rw, choice)
	if err != nil {
		fmt.Println("Error sending choice to server:", err)
		return
	}

	fileSize, fileHash, err := GetFileSizeAndHash(rw)
	if err != nil {
		fmt.Println("Error getting file size and hash:", err)
		return
	}

	err = ReceiveFile(rw, files[choice-1], fileSize, fileHash)
	if err != nil {
		fmt.Println("Error receiving file:", err)
		return
	}

	fmt.Println("File received and verified successfully")
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

func GetUserChoice(numFiles int) (int, error) {
	var choice int
	for {
		fmt.Print("Enter the number of the file to download: ")
		_, err := fmt.Scan(&choice)
		if err != nil || choice < 1 || choice > numFiles {
			fmt.Println("Invalid choice. Please enter a number between 1 and", numFiles)
			continue
		}
		break
	}
	return choice, nil
}

func SendChoice(rw *bufio.ReadWriter, choice int) error {
	_, err := rw.WriteString(strconv.Itoa(choice) + "\n")
	if err != nil {
		return err
	}
	return rw.Flush()
}

func GetFileSizeAndHash(rw *bufio.ReadWriter) (int64, string, error) {
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
