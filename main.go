package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.lsp.dev/jsonrpc2"
	"go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"
	"java-mini-ls-go/server"
	"net"
	"os"
)

var (
	listen = kingpin.Flag("listen", "Start the server and listen for incoming connections, "+
		"rather than trying to connect to an existing socket.").Bool()
	port = kingpin.Flag("socket", "The port to use").Default("9257").Int()
)

func main() {
	kingpin.Parse()

	logConfig := zap.NewDevelopmentConfig()
	logConfig.OutputPaths = []string{"stdout"}
	logConfig.ErrorOutputPaths = []string{"stdout"}

	logger, err := logConfig.Build()
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	defer func(logger *zap.Logger) {
		// flushes buffer, if any
		err := logger.Sync()
		if err != nil {
			fmt.Println("Error flushing logs: ", err)
		}
	}(logger)

	// Background context as parent of all LSP stuff
	ctx := context.Background()
	// Add logger as value on context
	ctx = context.WithValue(ctx, struct{}{}, logger)
	// Cancel it on program exit
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := runServer(ctx, logger); err != nil {
		logger.Error("failed to run server", zap.Error(err))
		os.Exit(1)
	}
}

func runServer(ctx context.Context, logger *zap.Logger) error {
	addr := fmt.Sprintf("127.0.0.1:%d", *port)

	if listen != nil && *listen {
		return runServerListen(ctx, logger, addr)
	} else {
		return runServerConnect(ctx, logger, addr)
	}
}

func runServerListen(ctx context.Context, logger *zap.Logger, addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrapf(err, "Error listening on %s", addr)
	}

	for {
		// Listen for connections
		conn, err := ln.Accept()
		if err != nil {
			return errors.Wrap(err, "Error accepting connection")
		}
		// Spin up a new language server instance for the connection
		go func() {
			server.RunServer(ctx, logger, jsonrpc2.NewStream(conn))
		}()
	}
}

func runServerConnect(ctx context.Context, logger *zap.Logger, addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return errors.Wrapf(err, "Error connecting to %s", addr)
	}

	_, con, _ := server.RunServer(ctx, logger, jsonrpc2.NewStream(conn))

	// Wait for connection to close
	doneChan := con.Done()
	<-doneChan

	logger.Info("Connection closed, server shutting down.")
	return nil
}
