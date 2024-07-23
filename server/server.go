package main

import (
	"crypto/rand"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"rcunov/net-transfer/utils"
	"strings"
	"time"
)

// Server configuration options
const (
	certFile = "server.pem"
	keyFile  = "server.key"
	port     = "6600"
)

const menu = "--- Menu: ---\n1. HELLO\n2. TIME\n3. EXIT\nEnter your choice: \f"

// Logging variables
var (
	level    = flag.String("loglevel", "info", "Set the logging level (debug, info, warn, error)")
	logger   *slog.Logger
	logLevel slog.Level
)

// StartServer starts listening on the assigned port using TLS with the provided certificate and private key.
func StartServer(port string, certFile string, keyFile string) (listener net.Listener, err error) {
	cert, err := utils.LoadCert(certFile, keyFile)
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

// HandleConnection performs the main server logic on an incoming connection.
func HandleConnection(conn net.Conn) {
	defer conn.Close()
	logMsg := fmt.Sprintf("client connected over TLS from %v", conn.RemoteAddr())
	logger.Info(logMsg)

	rw := utils.CreateReadWriter(conn)

	for { // Main logic loop
		rw.WriteString(menu)
		rw.Flush() // Have to remember to flush writes - write buffer is 4kb
		clientInput, err := rw.ReadString('\n')
		if err != nil {
			logger.Error("Connection closed")
			return
		}

		clientInput = strings.TrimSpace(clientInput)
		var response string

		switch clientInput {
		case "1":
			logMsg := fmt.Sprintf("received command from %v: %s", conn.RemoteAddr(), clientInput)
			logger.Debug(logMsg)
			response = "Hello, client!\n"
		case "2":
			logMsg := fmt.Sprintf("received command from %v: %s", conn.RemoteAddr(), clientInput)
			logger.Debug(logMsg)
			response = fmt.Sprintf("The current time is: %s\n", time.Now().Format(time.RFC1123))
		case "3":
			response = "Goodbye!\f"
			rw.WriteString(response)
			rw.Flush()
			logMsg := fmt.Sprintf("closing connection to %v", conn.RemoteAddr())
			logger.Info(logMsg)
			return
		default:
			logMsg := fmt.Sprintf("received invalid command from %v: %s", conn.RemoteAddr(), clientInput)
			logger.Debug(logMsg)
			response = "Invalid choice. Please select a valid option using the corresponding number.\n"
		}

		rw.WriteString(response)
		rw.Flush()
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

	logger = slog.New(slog.NewJSONHandler(os.Stderr, opts))
}

func main() {
	server, err := StartServer(port, certFile, keyFile)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer server.Close()

	debugMsg := fmt.Sprintf("server listening on port %v", port)
	logger.Info(debugMsg)

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Println(err)
			logger.Error(err.Error())
			continue
		}
		go HandleConnection(conn)
	}
}
