package utils

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"time"
)

// GenerateCert creates a new TLS certificate to encrypt network communications.
func GenerateCert() (cert tls.Certificate, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Generate a pem block with the private key
	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	// Generate random serial number for cert
	serial, err := rand.Int(rand.Reader, big.NewInt(123456))
	if err != nil {
		return tls.Certificate{}, err
	}

	// Generate the cert
	template := x509.Certificate{
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(5, 0, 0),
		// Have to generate a different serial number each execution
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: "rcunov/net-transfer",
		},
		BasicConstraintsValid: true,
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Generate a pem block with the cert
	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	// Parse the cert so we can use it
	cert, err = tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return tls.Certificate{}, err
	}

	return cert, err
}

func IsValidPort(port string) bool {
	p, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	return p > 1024 && p <= 65535
}

// GetCurrentDirectoryFiles returns a list of file names in the current directory.
func GetCurrentDirectoryFiles() ([]string, error) {
	files, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}

	fileNames := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}
	return fileNames, nil
}

// CalculateFileHash calculates the SHA-256 hash of the specified file.
func CalculateFileHash(fileName string) (hash string, err error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileHash := sha256.New()
	_, err = io.Copy(fileHash, file)
	if err != nil {
		return "", err
	}

	hash = hex.EncodeToString(fileHash.Sum(nil))
	return hash, nil
}

func CalculateFileSizeAndHash(fileName string) (fileSize int64, fileHash string, err error) {
	file, err := os.Open(fileName)
	if err != nil {
		return 0, "", err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return 0, "", err
	}
	fileSize = fileInfo.Size()

	fileHash, err = CalculateFileHash(fileName)
	if err != nil {
		return 0, "", err
	}

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
	calculatedHash, err := CalculateFileHash(file.Name())
	if err != nil {
		return err
	}

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
