package server

import (
	"context"
	"fmt"
	"go.lsp.dev/protocol"
)

type DiagnosticsPublisher interface {
	PublishDiagnostics(j *JavaLS, textDocument protocol.TextDocumentItem, diagnostics []protocol.Diagnostic)
}

type RealDiagnosticsPublisher struct{}

func (rdp *RealDiagnosticsPublisher) PublishDiagnostics(j *JavaLS, textDocument protocol.TextDocumentItem, diagnostics []protocol.Diagnostic) {
	// publish diagnostics in a separate goroutine
	go func(diagnostics []protocol.Diagnostic) {
		params := &protocol.PublishDiagnosticsParams{
			URI:         textDocument.URI,
			Version:     uint32(textDocument.Version),
			Diagnostics: diagnostics,
		}

		err := j.client.PublishDiagnostics(context.Background(), params)
		if err != nil {
			j.log.Error(fmt.Sprintf("Error publishing diagnostics: %v", err))
		}
	}(diagnostics)
}
