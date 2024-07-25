package main

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"rcunov/net-transfer/utils"
	"strconv"
)

// StartServer starts listening on the assigned port using TLS with the provided certificate and private key.
func StartServer(port string) (listener net.Listener, err error) {
	if !utils.IsValidPort(port) {
		return nil, fmt.Errorf("invalid port specified: %v. should be 1025-65535", port)
	}

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
