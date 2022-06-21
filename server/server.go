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
	ctx context.Context
	log *zap.Logger

	documentTextCache *util.SyncMap[string, protocol.TextDocumentItem]
	//parsedDocumentCache *util.SyncMap[string, antlr.Tree]
	symbols *util.SyncMap[string, []*parse.CodeSymbol]
}

func RunServer(ctx context.Context, logger *zap.Logger, stream jsonrpc2.Stream) (context.Context, jsonrpc2.Conn, protocol.Client) {
	jls := &JavaLS{
		ctx:               ctx,
		log:               logger,
		documentTextCache: util.NewSyncMap[string, protocol.TextDocumentItem](),
		symbols:           util.NewSyncMap[string, []*parse.CodeSymbol](),
		//parsedDocumentCache: util.NewSyncMap[string, antlr.Tree](),
	}

	return protocol.NewServer(ctx, jls, stream, jls.log.Named("client"))
}

func (j *JavaLS) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	j.log.Info("Initialize")
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
	j.parseTextDocument(params.TextDocument)
	return nil
}

func (j *JavaLS) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
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

	parsed := parse.Parse(textDocument.Text)
	//j.parsedDocumentCache.Set(uriString, parsed)

	symbols := parse.FindSymbols(parsed)
	fmt.Println("symbols:", symbols)
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

func (j *JavaLS) DocumentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) ([]interface{}, error) {
	ret := []interface{}{}

	symbols, _ := j.symbols.Get((string)(params.TextDocument.URI))

	if symbols != nil {
		docSymbols := convertToDocumentSymbols(symbols)
		for _, ds := range docSymbols {
			ret = append(ret, ds)
		}
	}

	fmt.Println("ret: ", ret)

	return ret, nil
}
