package server

import (
	"context"
	"fmt"
	"go.lsp.dev/uri"
	"golang.org/x/exp/slices"
	"java-mini-ls-go/parse"
	"java-mini-ls-go/parse/loc"
	"java-mini-ls-go/parse/sym"
	"java-mini-ls-go/parse/typ"
	"java-mini-ls-go/parse/typecheck"
	"java-mini-ls-go/util"

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
	symbols           *util.SyncMap[string, []*sym.CodeSymbol]
	scopes            *util.SyncMap[string, *typecheck.TypeCheckingScope]
	defUsages         *util.SyncMap[string, *typecheck.DefinitionsUsagesLookup]
	builtinTypes      *typ.TypeMap

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
		symbols:              util.NewSyncMap[string, []*sym.CodeSymbol](),
		scopes:               util.NewSyncMap[string, *typecheck.TypeCheckingScope](),
		defUsages:            util.NewSyncMap[string, *typecheck.DefinitionsUsagesLookup](),
		builtinTypes:         typ.NewTypeMap(),
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
			ReferencesProvider:     true,
			DefinitionProvider:     true,
			CompletionProvider: &protocol.CompletionOptions{
				ResolveProvider:   false,
				TriggerCharacters: []string{"."},
			},
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

	symbols := sym.FindSymbols(parsed)
	j.symbols.Set(uriString, symbols)

	typeCheckingResult := typecheck.CheckTypes(j.log, uriString, int(textDocument.Version), parsed, j.builtinTypes)
	j.scopes.Set(uriString, typeCheckingResult.RootScope)
	j.defUsages.Set(uriString, typeCheckingResult.DefUsagesLookup)

	typeErrors := typeCheckingResult.TypeErrors
	diagnostics := util.CombineSlices(
		util.Map(syntaxErrors, func(se parse.SyntaxError) protocol.Diagnostic { return se.ToDiagnostic() }),
		util.Map(typeErrors, func(se typecheck.TypeError) protocol.Diagnostic { return se.ToDiagnostic() }),
	)

	j.diagnosticsPublisher.PublishDiagnostics(j, textDocument, diagnostics)
}

var symbolTypeMap = map[sym.CodeSymbolType]protocol.SymbolKind{
	sym.CodeSymbolClass:       protocol.SymbolKindClass,
	sym.CodeSymbolConstant:    protocol.SymbolKindConstant,
	sym.CodeSymbolConstructor: protocol.SymbolKindConstructor,
	sym.CodeSymbolEnum:        protocol.SymbolKindEnum,
	sym.CodeSymbolEnumMember:  protocol.SymbolKindEnumMember,
	sym.CodeSymbolInterface:   protocol.SymbolKindInterface,
	sym.CodeSymbolMethod:      protocol.SymbolKindMethod,
	sym.CodeSymbolPackage:     protocol.SymbolKindPackage,
	sym.CodeSymbolVariable:    protocol.SymbolKindVariable,
}

func convertToDocumentSymbols(codeSymbols []*sym.CodeSymbol) []protocol.DocumentSymbol {
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
	lookup, ok := j.defUsages.Get(string(params.TextDocument.URI))
	if !ok {
		return nil, nil
	}

	symbol := lookup.Lookup(loc.FileLocation{
		// Note: the +1 is convert from 0-based line numbers (LSP) to 1-based line numbers (this project)
		Line:      int(params.Position.Line) + 1,
		Character: int(params.Position.Character),
	})
	if symbol == nil {
		return nil, nil
	}

	// For now just return the variable name + type
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: fmt.Sprintf("**%s** %s", symbol.ShortName(), symbol.FullName()),
		},
		Range: nil,
	}, nil
}

// NOTE: line is 0-based here (LSP style)
// Will probably be used for auto-completion
//nolint:unused
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

func (j *JavaLS) References(_ context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	lookup, ok := j.defUsages.Get(string(params.TextDocument.URI))
	if !ok {
		return nil, nil
	}

	symbol := lookup.Lookup(loc.FileLocation{
		// Note: the +1 is convert from 0-based line numbers (LSP) to 1-based line numbers (this project)
		Line:      int(params.Position.Line) + 1,
		Character: int(params.Position.Character),
	})
	if symbol == nil {
		return nil, nil
	}

	return util.Map(symbol.GetUsages(), codeLocationToLSPLocation), nil
}

func codeLocationToLSPLocation(loca loc.CodeLocation) protocol.Location {
	return protocol.Location{
		URI:   uri.New(loca.FileUri),
		Range: loc.BoundsToRange(loca.Loc),
	}
}

func (j *JavaLS) Definition(_ context.Context, params *protocol.DefinitionParams) ([]protocol.Location, error) {
	lookup, ok := j.defUsages.Get(string(params.TextDocument.URI))
	if !ok {
		return nil, nil
	}

	symbol := lookup.Lookup(loc.FileLocation{
		// Note: the +1 is convert from 0-based line numbers (LSP) to 1-based line numbers (this project)
		Line:      int(params.Position.Line) + 1,
		Character: int(params.Position.Character),
	})
	if symbol == nil || symbol.GetDefinition() == nil {
		return nil, nil
	}

	return []protocol.Location{codeLocationToLSPLocation(*symbol.GetDefinition())}, nil
}

func (j *JavaLS) Completion(_ context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	textOnLine, err := j.getTextOnLine(string(params.TextDocument.URI), int(params.Position.Line))
	if err != nil {
		return nil, err
	}
	// go backwards from the current character to see if this was triggered by a dot `.`, and if so,
	// what's the symbol on the left of the dot
	dotIdx := -1
	startIdx := int(params.Position.Character) - 1
	if startIdx >= len(textOnLine) {
		startIdx = len(textOnLine) - 1
	}
	for i := startIdx; i >= 0; i-- {
		ch := textOnLine[i]
		if ch == '.' {
			dotIdx = i
			break
		}
		if !isAlphaNumeric(ch) && !isWhitespace(ch) {
			break
		}
	}
	if dotIdx != -1 {
		j.log.Info("found a dot")

		// We've got a dot. What is the symbol on the left of the dot?
		defUsages, ok := j.defUsages.Get(string(params.TextDocument.URI))
		if ok {
			leftOfDot := defUsages.Lookup(loc.FileLocation{
				Line:      int(params.Position.Line + 1),
				Character: dotIdx,
			})
			if leftOfDot != nil {
				allMembers := leftOfDot.GetType().AllMembers()
				j.log.Info(fmt.Sprintf("Auto-complete dot items: %d", len(allMembers)))
				return symbolsToCompletionList(allMembers), nil
			}
		}
	}

	fileScopes, ok := j.scopes.Get(string(params.TextDocument.URI))
	if ok {
		scope := fileScopes.LookupScopeFor(loc.FileLocation{
			Line:      int(params.Position.Line + 1),
			Character: int(params.Position.Character),
		})
		symbols := scope.AllSymbols()
		j.log.Info(fmt.Sprintf("Auto-complete non-dot items: %d", len(symbols)))
		return symbolsToCompletionList(symbols), nil
	}

	return nil, nil
}

func isAlphaNumeric(ch uint8) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_'
}

var whitespaceChars = []uint8{
	'\n', '\r', '\t', '\v', ' ',
}

func isWhitespace(ch uint8) bool {
	return slices.Index(whitespaceChars, ch) != -1
}

func symbolsToCompletionList(symbols []typ.JavaSymbol) *protocol.CompletionList {
	completions := util.Map(symbols, func(s typ.JavaSymbol) protocol.CompletionItem {
		return protocol.CompletionItem{Label: s.ShortName()} //nolint:exhaustruct
	})
	return &protocol.CompletionList{
		IsIncomplete: false,
		Items:        completions,
	}
}
