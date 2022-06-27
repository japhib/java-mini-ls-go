package server

import (
	"context"
	"fmt"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
	"java-mini-ls-go/parse"
	"java-mini-ls-go/util"
)

// Runtime check to ensure JavaLS implements interface
var _ protocol.Server = (*JavaLS)(nil)

type JavaLS struct {
	ctx    context.Context
	log    *zap.Logger
	client protocol.Client

	documentTextCache *util.SyncMap[string, protocol.TextDocumentItem]
	symbols           *util.SyncMap[string, []*parse.CodeSymbol]
	builtinTypes      map[string]*parse.JavaType

	// Options
	ReadStdlibTypes bool
}

func NewServer(ctx context.Context, logger *zap.Logger) *JavaLS {
	return &JavaLS{
		ctx:               ctx,
		log:               logger,
		documentTextCache: util.NewSyncMap[string, protocol.TextDocumentItem](),
		symbols:           util.NewSyncMap[string, []*parse.CodeSymbol](),
		builtinTypes:      make(map[string]*parse.JavaType),
	}
}

func RunServer(ctx context.Context, logger *zap.Logger, stream jsonrpc2.Stream) (context.Context, jsonrpc2.Conn, protocol.Client) {
	jls := NewServer(ctx, logger)
	jls.ReadStdlibTypes = true

	ctx, conn, client := protocol.NewServer(ctx, jls, stream, jls.log.Named("client"))
	jls.client = client

	return ctx, conn, client
}

func (j *JavaLS) Initialize(_ context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	j.log.Info("Initialize")

	if j.ReadStdlibTypes {
		var err error
		j.builtinTypes, err = parse.LoadBuiltinTypes()
		if err != nil {
			j.log.Error(err.Error())
			return nil, err
		}
	}

	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: protocol.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    protocol.TextDocumentSyncKindFull,
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

func (j *JavaLS) Initialized(_ context.Context, params *protocol.InitializedParams) error {
	j.log.Info("Initialized")
	return nil
}

func (j *JavaLS) Shutdown(_ context.Context) error {
	return nil
}

func (j *JavaLS) Exit(_ context.Context) error {
	return nil
}

func (j *JavaLS) DidOpen(_ context.Context, params *protocol.DidOpenTextDocumentParams) error {
	j.log.Info(fmt.Sprintf("DidOpen %s", params.TextDocument.URI))
	j.parseTextDocument(params.TextDocument)
	return nil
}

func (j *JavaLS) DidChange(_ context.Context, params *protocol.DidChangeTextDocumentParams) error {
	j.log.Info(fmt.Sprintf("DidChange %s", params.TextDocument.URI))
	item := protocol.TextDocumentItem{
		URI:     params.TextDocument.URI,
		Version: params.TextDocument.Version,
		Text:    params.ContentChanges[0].Text,
		// NOTE: language ID not set here
	}
	j.parseTextDocument(item)
	return nil
}

func (j *JavaLS) parseTextDocument(textDocument protocol.TextDocumentItem) {
	uriString := (string)(textDocument.URI)
	j.documentTextCache.Set(uriString, textDocument)

	parsed, errors := parse.Parse(textDocument.Text)

	if len(errors) > 0 {
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

	symbols := parse.FindSymbols(parsed)
	j.symbols.Set(uriString, symbols)
}

var symbolTypeMap = map[parse.CodeSymbolType]protocol.SymbolKind{
	parse.CodeSymbolClass:       protocol.SymbolKindClass,
	parse.CodeSymbolConstant:    protocol.SymbolKindConstant,
	parse.CodeSymbolConstructor: protocol.SymbolKindConstructor,
	parse.CodeSymbolEnum:        protocol.SymbolKindEnum,
	parse.CodeSymbolEnumMember:  protocol.SymbolKindEnumMember,
	parse.CodeSymbolInterface:   protocol.SymbolKindInterface,
	parse.CodeSymbolMethod:      protocol.SymbolKindMethod,
	parse.CodeSymbolPackage:     protocol.SymbolKindPackage,
	parse.CodeSymbolVariable:    protocol.SymbolKindVariable,
}

func convertToDocumentSymbols(codeSymbols []*parse.CodeSymbol) []protocol.DocumentSymbol {
	ret := make([]protocol.DocumentSymbol, 0, len(codeSymbols))

	for _, s := range codeSymbols {
		rrange := parse.BoundsToRange(s.Bounds)
		documentSymbol := protocol.DocumentSymbol{
			Name:           s.Name,
			Detail:         s.Detail,
			Kind:           symbolTypeMap[s.Type],
			Range:          rrange,
			SelectionRange: rrange,
			Children:       convertToDocumentSymbols(s.Children),
		}
		ret = append(ret, documentSymbol)
	}

	return ret
}

func (j *JavaLS) DocumentSymbol(_ context.Context, params *protocol.DocumentSymbolParams) ([]interface{}, error) {
	ret := make([]interface{}, 0)

	symbols, _ := j.symbols.Get((string)(params.TextDocument.URI))

	if symbols != nil {
		docSymbols := convertToDocumentSymbols(symbols)
		ret = make([]interface{}, 0, len(docSymbols))
		for _, ds := range docSymbols {
			ret = append(ret, ds)
		}
	}

	return ret, nil
}
