package server

import (
	"context"
	"fmt"
	"java-mini-ls-go/parse"
	"java-mini-ls-go/parse/loc"
	"java-mini-ls-go/parse/symbol"
	"java-mini-ls-go/parse/typ"
	"java-mini-ls-go/parse/typecheck"
	"java-mini-ls-go/util"

	"github.com/pkg/errors"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

// Runtime check to ensure JavaLS implements interface
var _ protocol.Server = (*JavaLS)(nil)

type JavaLS struct {
	ctx    context.Context
	log    *zap.Logger
	client protocol.Client

	documentTextCache *util.SyncMap[string, protocol.TextDocumentItem]
	symbols           *util.SyncMap[string, []*symbol.CodeSymbol]
	scopes            *util.SyncMap[string, typecheck.TypeCheckingScope]
	defUsages         *util.SyncMap[string, *typecheck.DefinitionsUsagesLookup]
	builtinTypes      map[string]*typ.JavaType

	// Dependencies that can be mocked for testing
	diagnosticsPublisher DiagnosticsPublisher

	// Options
	ReadStdlibTypes bool
}

func NewServer(ctx context.Context, logger *zap.Logger) *JavaLS {
	return &JavaLS{
		ctx:                  ctx,
		log:                  logger,
		client:               nil,
		documentTextCache:    util.NewSyncMap[string, protocol.TextDocumentItem](),
		symbols:              util.NewSyncMap[string, []*symbol.CodeSymbol](),
		scopes:               util.NewSyncMap[string, typecheck.TypeCheckingScope](),
		defUsages:            util.NewSyncMap[string, *typecheck.DefinitionsUsagesLookup](),
		builtinTypes:         make(map[string]*typ.JavaType),
		diagnosticsPublisher: &RealDiagnosticsPublisher{},
		ReadStdlibTypes:      false,
	}
}

func RunServer(ctx context.Context, logger *zap.Logger, stream jsonrpc2.Stream) (context.Context, jsonrpc2.Conn, protocol.Client) {
	jls := NewServer(ctx, logger)
	jls.ReadStdlibTypes = true

	ctx, conn, client := protocol.NewServer(ctx, jls, stream, jls.log.Named("client"))
	jls.client = client

	return ctx, conn, client
}

//nolint:exhaustruct
func (j *JavaLS) Initialize(_ context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	j.log.Info("Initialize")

	var err error
	j.builtinTypes, err = typ.LoadBuiltinTypes()
	if err != nil {
		j.log.Error(err.Error())
		return nil, err
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
			HoverProvider:          true,
		},
		ServerInfo: nil,
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
		URI:        params.TextDocument.URI,
		Version:    params.TextDocument.Version,
		Text:       params.ContentChanges[0].Text,
		LanguageID: "java",
	}
	j.parseTextDocument(item)
	return nil
}

func (j *JavaLS) parseTextDocument(textDocument protocol.TextDocumentItem) {
	uriString := string(textDocument.URI)
	j.documentTextCache.Set(uriString, textDocument)

	parsed, syntaxErrors := parse.Parse(textDocument.Text)

	symbols := symbol.FindSymbols(parsed)
	j.symbols.Set(uriString, symbols)

	userTypes := typecheck.GatherTypes(uriString, parsed, j.builtinTypes)
	typeCheckingResult := typecheck.CheckTypes(j.log, parsed, uriString, userTypes, j.builtinTypes)
	j.scopes.Set(uriString, typeCheckingResult.RootScope)
	j.defUsages.Set(uriString, typeCheckingResult.DefUsagesLookup)

	typeErrors := typeCheckingResult.TypeErrors
	diagnostics := util.CombineSlices(
		util.Map(syntaxErrors, func(se parse.SyntaxError) protocol.Diagnostic { return se.ToDiagnostic() }),
		util.Map(typeErrors, func(se typecheck.TypeError) protocol.Diagnostic { return se.ToDiagnostic() }),
	)

	j.diagnosticsPublisher.PublishDiagnostics(j, textDocument, diagnostics)
}

var symbolTypeMap = map[symbol.CodeSymbolType]protocol.SymbolKind{
	symbol.CodeSymbolClass:       protocol.SymbolKindClass,
	symbol.CodeSymbolConstant:    protocol.SymbolKindConstant,
	symbol.CodeSymbolConstructor: protocol.SymbolKindConstructor,
	symbol.CodeSymbolEnum:        protocol.SymbolKindEnum,
	symbol.CodeSymbolEnumMember:  protocol.SymbolKindEnumMember,
	symbol.CodeSymbolInterface:   protocol.SymbolKindInterface,
	symbol.CodeSymbolMethod:      protocol.SymbolKindMethod,
	symbol.CodeSymbolPackage:     protocol.SymbolKindPackage,
	symbol.CodeSymbolVariable:    protocol.SymbolKindVariable,
}

func convertToDocumentSymbols(codeSymbols []*symbol.CodeSymbol) []protocol.DocumentSymbol {
	ret := make([]protocol.DocumentSymbol, 0, len(codeSymbols))

	for _, s := range codeSymbols {
		rrange := loc.BoundsToRange(s.Bounds)
		documentSymbol := protocol.DocumentSymbol{
			Name:           s.Name,
			Detail:         s.Detail,
			Kind:           symbolTypeMap[s.Type],
			Tags:           nil,
			Deprecated:     false,
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

	symbols, _ := j.symbols.Get(string(params.TextDocument.URI))

	if symbols != nil {
		docSymbols := convertToDocumentSymbols(symbols)
		ret = make([]interface{}, 0, len(docSymbols))
		for _, ds := range docSymbols {
			ret = append(ret, ds)
		}
	}

	return ret, nil
}

func (j *JavaLS) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	textOnLine, err := j.getTextOnLine(string(params.TextDocument.URI), int(params.Position.Line))
	if textOnLine == "" {
		return nil, errors.Wrapf(err, "can't get document text on line %d", int(params.Position.Line))
	}

	// Check if it's a local
	lookup, ok := j.defUsages.Get(string(params.TextDocument.URI))
	if ok {
		symbol := lookup.Lookup(loc.FileLocation{
			Line:   int(params.Position.Line) + 1,
			Column: int(params.Position.Character),
		})
		if symbol != nil {
			// For now just return the variable name + type
			return &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.Markdown,
					Value: fmt.Sprintf("**%s** %s", symbol.ShortName(), symbol.FullName()),
				},
				Range: nil,
			}, nil
		}
	}

	return nil, nil
}

// NOTE: line is 0-based here (LSP style)
func (j *JavaLS) getTextOnLine(fileURI string, line int) (string, error) {
	text, ok := j.documentTextCache.Get(fileURI)
	if !ok {
		return "", fmt.Errorf("can't find document with uri: %s", fileURI)
	}

	// TODO use a different data structure than just a string so that this lookup isn't O(n)
	currLine := 0
	startIdx := 0
	foundStart := false
	endIdx := 0
	foundEnd := false

	if line == 0 {
		startIdx = -1
		foundStart = true
	}

	for currCharIdx := 0; currCharIdx < len(text.Text); currCharIdx++ {
		currChar := text.Text[currCharIdx]
		if currChar == '\n' {
			currLine++

			if currLine == line {
				foundStart = true
				startIdx = currCharIdx
			} else if currLine == line+1 {
				foundEnd = true
				endIdx = currCharIdx
				break
			}
		}
	}
	if !foundStart {
		return "", fmt.Errorf("can't find line %d, document only has %d lines total", line, currLine)
	}
	if !foundEnd {
		// it's the last line
		endIdx = len(text.Text)
	}

	return text.Text[startIdx+1 : endIdx], nil
}
