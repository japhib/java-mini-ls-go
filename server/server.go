package server

import (
	"context"
	"fmt"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
	"java-mini-ls-go/util"
)

// Runtime check to ensure JavaLS implements interface
var _ protocol.Server = (*JavaLS)(nil)

type JavaLS struct {
	ctx context.Context
	log *zap.Logger

	documentTextCache util.SyncMap[string, protocol.TextDocumentItem]
}

func RunServer(ctx context.Context, logger *zap.Logger, stream jsonrpc2.Stream) (context.Context, jsonrpc2.Conn, protocol.Client) {
	jls := &JavaLS{
		ctx: ctx,
		log: logger,
	}

	return protocol.NewServer(ctx, jls, stream, jls.log.Named("client"))
}

func (j *JavaLS) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	j.log.Info("Initialize")
	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: protocol.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    protocol.TextDocumentSyncKindIncremental,
			},
			Workspace: &protocol.ServerCapabilitiesWorkspace{
				WorkspaceFolders: &protocol.ServerCapabilitiesWorkspaceFolders{
					Supported:           true,
					ChangeNotifications: true,
				},
			},
			DocumentSymbolProvider: true,
		},
	}, nil
}

func (j *JavaLS) Initialized(ctx context.Context, params *protocol.InitializedParams) error {
	j.log.Info("Initialized")
	return nil
}

func (j *JavaLS) Shutdown(ctx context.Context) error {
	return nil
}

func (j *JavaLS) Exit(ctx context.Context) error {
	return nil
}

func (j *JavaLS) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	j.log.Info(fmt.Sprintf("DidOpen %s", params.TextDocument.URI))
	return nil
}

func (j *JavaLS) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	j.log.Info(fmt.Sprintf("DidChange %s", params.TextDocument.URI))
	return nil
}

func (j *JavaLS) parseTextDocument(textDocument protocol.TextDocumentItem) {
	j.documentTextCache.Set(textDocument.URI.Filename(), textDocument)
}
