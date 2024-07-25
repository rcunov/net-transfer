package main

import (
	"bufio"
	"errors"
	"fmt"
	"rcunov/net-transfer/utils"
	"strconv"
)

func HandleFileDownload(rw *bufio.ReadWriter, remoteAddr string) error {
	// files := []string{"file1.txt", "file2.txt", "file3.txt"}
	files, err := utils.GetCurrentDirectoryFiles()
	if err != nil {
		return err
	}

	err = SendNumberOfFiles(rw, len(files))
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
	fileSize, fileHash, err := utils.CalculateFileSizeAndHash(fileName)
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

	fmt.Printf("File %v sent successfully to %v\n", fileName, remoteAddr)
	return nil
}

func HandleFileUpload(rw *bufio.ReadWriter, remoteAddr string) error {
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

	err = utils.ReceiveFile(rw, fileName, fileSize, fileHash)
	if errors.As(err, &netErr) {
		return fmt.Errorf("client has disconnected")
	}
	if err != nil {
		return err
	}

	fmt.Printf("File %v received from %v and verified successfully\n", fileName, remoteAddr)
	return nil
}
