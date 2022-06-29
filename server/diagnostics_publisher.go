package server

import (
	"context"
	"fmt"
	"go.lsp.dev/protocol"
	"java-mini-ls-go/parse"
	"java-mini-ls-go/util"
)

type DiagnosticsPublisher interface {
	PublishDiagnostics(j *JavaLS, textDocument protocol.TextDocumentItem, errors []parse.SyntaxError)
}

type RealDiagnosticsPublisher struct{}

func (rdp *RealDiagnosticsPublisher) PublishDiagnostics(j *JavaLS, textDocument protocol.TextDocumentItem, errors []parse.SyntaxError) {
	// publish diagnostics in a separate goroutine
	go func(errors []parse.SyntaxError) {
		params := &protocol.PublishDiagnosticsParams{
			URI:         textDocument.URI,
			Version:     (uint32)(textDocument.Version),
			Diagnostics: util.Map(errors, func(se parse.SyntaxError) protocol.Diagnostic { return se.ToDiagnostic() }),
		}

		err := j.client.PublishDiagnostics(context.Background(), params)
		if err != nil {
			j.log.Error(fmt.Sprintf("Error publishing diagnostics: %v", err))
		}
	}(errors)
}
