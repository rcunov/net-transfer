package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"os"
	"rcunov/net-transfer/utils"
	"strings"
	"time"
)

// Server connection info
const (
	hostname = "localhost"
	port     = "6600"
)

var (
	level    = flag.String("loglevel", "info", "Set the logging level (debug, info, warn, error)")
	logger   *slog.Logger
	logLevel slog.Level
)

// GenerateCert creates a new TLS certificate to encrypt the connection to the server.
// If an error is returned, the certificate will be empty.
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
			CommonName: "net-transfer client",
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

// ConnectToServer initiates a TLS connection to the server at the provided hostname and port.
func ConnectToServer(tlsCert tls.Certificate, hostname string, port string) (conn net.Conn, err error) {
	logger.Debug("connecting to server at " + hostname)
	config := tls.Config{Certificates: []tls.Certificate{tlsCert}, InsecureSkipVerify: true}
	conn, err = tls.Dial("tcp", hostname+":"+port, &config)
	if err != nil {
		return nil, err
	}

	localPort := conn.LocalAddr().(*net.TCPAddr).Port
	logMsg := fmt.Sprintf("connection established. localhost:%v --> %v:%v", localPort, hostname, port)
	logger.Debug(logMsg)

	return conn, err
}

// ReadInput reads and validates input from user. When a user does not enter
// a selection but presses the enter key, ReadInput will prompt the user for a
// non-empty selection and re-read them the menu supplied by the server. Once a
// selection has been entered by the user, it will be sent to the server for validation.
func ReadInput(reader *bufio.Reader, menu string) (input string) {
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			logger.Debug("user input empty, prompting for valid selection")
			fmt.Print("\n--> Error! Please enter a selection.\n\n")
			fmt.Print(menu)
		} else {
			return input
		}
	}
}

// Set up logging
func init() {
	flag.Parse()

	switch *level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		fmt.Println("Invalid log level specified. Defaulting to info.")
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	logger = slog.New(slog.NewTextHandler(os.Stderr, opts))
}

func main() {
	tlsCert, err := GenerateCert()
	if err != nil {
		logMsg := fmt.Sprintf("cannot generate the certificate: %v", err.Error())
		logger.Error(logMsg)
		os.Exit(1)
	}

	conn, err := ConnectToServer(tlsCert, hostname, port)
	if err != nil {
		logMsg := fmt.Sprintf("cannot connect to server: %v", err.Error())
		logger.Error(logMsg)
		os.Exit(1)
	}
	defer conn.Close()

	rw := utils.CreateReadWriter(conn) // For reading and writing to server
	stdin := bufio.NewReader(os.Stdin) // For reading input from user

	fmt.Println()
	for {
		menu, err := rw.ReadString('\f') // Form feed is escape sequence
		if err != nil {
			logMsg := fmt.Sprintf("error reading menu from server: %v", err.Error())
			logger.Error(logMsg)
			return
		}
		menu = fmt.Sprintf(menu[0:len(menu)-2] + " ") // Remove trailing escape sequence
		fmt.Print(menu)

		input := ReadInput(stdin, menu)
		fmt.Println()

		rw.WriteString(input + "\n")
		rw.Flush()

		message, err := rw.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Print("--> Server closed the connection.\n\n")
				break
			}
			var netErr *net.OpError
			if errors.As(err, &netErr) {
				fmt.Print("--> Server has shut down.\n\n")
				break
			}
			logMsg := fmt.Sprintf("error reading from server: %v", err.Error())
			logger.Error(logMsg)
			break
		}

		fmt.Printf("--> Server response: %s\n", message)
	}
}
