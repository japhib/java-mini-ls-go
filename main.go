package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.lsp.dev/jsonrpc2"
	"go.uber.org/zap"
	"java-mini-ls-go/server"
	"net"
	"os"
)

// The port to listen on
const port = 9257

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	defer func(logger *zap.Logger) {
		// flushes buffer, if any
		err := logger.Sync()
		if err != nil {
			fmt.Println(err)
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
	addr := fmt.Sprintf("127.0.0.1:%d", port)
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
