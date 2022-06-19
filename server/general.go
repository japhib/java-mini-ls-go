package server

import (
	"context"
	"go.lsp.dev/protocol"
)

// This file contains initialize/shutdown type calls

func (j *JavaLS) initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
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
