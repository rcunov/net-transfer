package main

import (
	"bufio"
	"fmt"
	"rcunov/net-transfer/utils"
	"strconv"
)

func HandleFileDownload(rw *bufio.ReadWriter) error {
	numFiles, err := GetNumberOfFiles(rw)
	if err != nil {
		return err
	}

	files, err := GetListOfFiles(rw, numFiles)
	if err != nil {
		return err
	}

	DisplayFiles(files)

	selection, err := GetUserSelection(len(files))
	if err != nil {
		return err
	}

	err = SendSelection(rw, selection)
	if err != nil {
		return err
	}

	fileSize, fileHash, err := ReceiveFileSizeAndHash(rw)
	if err != nil {
		return err
	}

	err = utils.ReceiveFile(rw, files[selection-1], fileSize, fileHash)
	if err != nil {
		return err
	}

	fmt.Println("File received and verified successfully")
	return nil
}

func HandleFileUpload(rw *bufio.ReadWriter) error {
	fmt.Print("Enter the file name to upload: ")
	var fileName string
	_, err := fmt.Scan(&fileName)
	if err != nil {
		return err
	}

	fileSize, fileHash, err := utils.CalculateFileSizeAndHash(fileName)
	if err != nil {
		return err
	}

	_, err = rw.WriteString(fileName + "\n")
	if err != nil {
		return err
	}
	rw.Flush()

	_, err = rw.WriteString(strconv.FormatInt(fileSize, 10) + "\n")
	if err != nil {
		return err
	}
	rw.Flush()

	fmt.Println("Awaiting approval from server...")
	approval, err := rw.ReadString('\n')
	if err != nil {
		return err
	}
	approval = approval[:len(approval)-1]

	if approval != "yes" {
		fmt.Println("File upload rejected by server.")
		return nil
	}
	fmt.Println("File upload approved!")

	_, err = rw.WriteString(fileHash + "\n")
	if err != nil {
		return err
	}
	rw.Flush()

	err = SendFile(rw, fileName)
	if err != nil {
		return err
	}

	fmt.Println("File uploaded successfully")
	return nil
}
