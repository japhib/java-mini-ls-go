package server

import (
	"context"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

// Runtime check to ensure JavaLS implements interface
var _ protocol.Server = (*JavaLS)(nil)

type JavaLS struct {
	ctx context.Context
	log *zap.Logger
}

func RunServer(ctx context.Context, logger *zap.Logger, stream jsonrpc2.Stream) {
	jls := &JavaLS{
		ctx: ctx,
		log: logger,
	}

	protocol.NewServer(ctx, jls, stream, jls.log.Named("client"))
}

func (j *JavaLS) Run() {

}

// Note that the implementations of specific LSP functions are farmed out to other
// files, in order to prevent this file from getting insanely big.

func (j *JavaLS) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	return j.initialize(ctx, params)
}

func (j *JavaLS) Initialized(ctx context.Context, params *protocol.InitializedParams) error {
	panic("unimplemented")
}

func (j *JavaLS) Shutdown(ctx context.Context) error {
	panic("unimplemented")
}

func (j *JavaLS) Exit(ctx context.Context) error {
	panic("unimplemented")
}

func (j *JavaLS) WorkDoneProgressCancel(ctx context.Context, params *protocol.WorkDoneProgressCancelParams) error {
	panic("unimplemented")
}

func (j *JavaLS) LogTrace(ctx context.Context, params *protocol.LogTraceParams) error {
	panic("unimplemented")
}

func (j *JavaLS) SetTrace(ctx context.Context, params *protocol.SetTraceParams) error {
	panic("unimplemented")
}

func (j *JavaLS) CodeAction(ctx context.Context, params *protocol.CodeActionParams) ([]protocol.CodeAction, error) {
	panic("unimplemented")
}

func (j *JavaLS) CodeLens(ctx context.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	panic("unimplemented")
}

func (j *JavaLS) CodeLensResolve(ctx context.Context, params *protocol.CodeLens) (*protocol.CodeLens, error) {
	panic("unimplemented")
}

func (j *JavaLS) ColorPresentation(ctx context.Context, params *protocol.ColorPresentationParams) ([]protocol.ColorPresentation, error) {
	panic("unimplemented")
}

func (j *JavaLS) Completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	panic("unimplemented")
}

func (j *JavaLS) CompletionResolve(ctx context.Context, params *protocol.CompletionItem) (*protocol.CompletionItem, error) {
	panic("unimplemented")
}

func (j *JavaLS) Declaration(ctx context.Context, params *protocol.DeclarationParams) ([]protocol.Location, error) {
	panic("unimplemented")
}

func (j *JavaLS) Definition(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.Location, error) {
	panic("unimplemented")
}

func (j *JavaLS) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	panic("unimplemented")
}

func (j *JavaLS) DidChangeConfiguration(ctx context.Context, params *protocol.DidChangeConfigurationParams) error {
	panic("unimplemented")
}

func (j *JavaLS) DidChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
	panic("unimplemented")
}

func (j *JavaLS) DidChangeWorkspaceFolders(ctx context.Context, params *protocol.DidChangeWorkspaceFoldersParams) error {
	panic("unimplemented")
}

func (j *JavaLS) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	panic("unimplemented")
}

func (j *JavaLS) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	panic("unimplemented")
}

func (j *JavaLS) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
	panic("unimplemented")
}

func (j *JavaLS) DocumentColor(ctx context.Context, params *protocol.DocumentColorParams) ([]protocol.ColorInformation, error) {
	panic("unimplemented")
}

func (j *JavaLS) DocumentHighlight(ctx context.Context, params *protocol.DocumentHighlightParams) ([]protocol.DocumentHighlight, error) {
	panic("unimplemented")
}

func (j *JavaLS) DocumentLink(ctx context.Context, params *protocol.DocumentLinkParams) ([]protocol.DocumentLink, error) {
	panic("unimplemented")
}

func (j *JavaLS) DocumentLinkResolve(ctx context.Context, params *protocol.DocumentLink) (*protocol.DocumentLink, error) {
	panic("unimplemented")
}

func (j *JavaLS) DocumentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) ([]interface{}, error) {
	panic("unimplemented")
}

func (j *JavaLS) ExecuteCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (interface{}, error) {
	panic("unimplemented")
}

func (j *JavaLS) FoldingRanges(ctx context.Context, params *protocol.FoldingRangeParams) ([]protocol.FoldingRange, error) {
	panic("unimplemented")
}

func (j *JavaLS) Formatting(ctx context.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	panic("unimplemented")
}

func (j *JavaLS) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	panic("unimplemented")
}

func (j *JavaLS) Implementation(ctx context.Context, params *protocol.ImplementationParams) ([]protocol.Location, error) {
	panic("unimplemented")
}

func (j *JavaLS) OnTypeFormatting(ctx context.Context, params *protocol.DocumentOnTypeFormattingParams) ([]protocol.TextEdit, error) {
	panic("unimplemented")
}

func (j *JavaLS) PrepareRename(ctx context.Context, params *protocol.PrepareRenameParams) (*protocol.Range, error) {
	panic("unimplemented")
}

func (j *JavaLS) RangeFormatting(ctx context.Context, params *protocol.DocumentRangeFormattingParams) ([]protocol.TextEdit, error) {
	panic("unimplemented")
}

func (j *JavaLS) References(ctx context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	panic("unimplemented")
}

func (j *JavaLS) Rename(ctx context.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	panic("unimplemented")
}

func (j *JavaLS) SignatureHelp(ctx context.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	panic("unimplemented")
}

func (j *JavaLS) Symbols(ctx context.Context, params *protocol.WorkspaceSymbolParams) ([]protocol.SymbolInformation, error) {
	panic("unimplemented")
}

func (j *JavaLS) TypeDefinition(ctx context.Context, params *protocol.TypeDefinitionParams) ([]protocol.Location, error) {
	panic("unimplemented")
}

func (j *JavaLS) WillSave(ctx context.Context, params *protocol.WillSaveTextDocumentParams) error {
	panic("unimplemented")
}

func (j *JavaLS) WillSaveWaitUntil(ctx context.Context, params *protocol.WillSaveTextDocumentParams) ([]protocol.TextEdit, error) {
	panic("unimplemented")
}

func (j *JavaLS) ShowDocument(ctx context.Context, params *protocol.ShowDocumentParams) (*protocol.ShowDocumentResult, error) {
	panic("unimplemented")
}

func (j *JavaLS) WillCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) (*protocol.WorkspaceEdit, error) {
	panic("unimplemented")
}

func (j *JavaLS) DidCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) error {
	panic("unimplemented")
}

func (j *JavaLS) WillRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) (*protocol.WorkspaceEdit, error) {
	panic("unimplemented")
}

func (j *JavaLS) DidRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) error {
	panic("unimplemented")
}

func (j *JavaLS) WillDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) (*protocol.WorkspaceEdit, error) {
	panic("unimplemented")
}

func (j *JavaLS) DidDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) error {
	panic("unimplemented")
}

func (j *JavaLS) CodeLensRefresh(ctx context.Context) error {
	panic("unimplemented")
}

func (j *JavaLS) PrepareCallHierarchy(ctx context.Context, params *protocol.CallHierarchyPrepareParams) ([]protocol.CallHierarchyItem, error) {
	panic("unimplemented")
}

func (j *JavaLS) IncomingCalls(ctx context.Context, params *protocol.CallHierarchyIncomingCallsParams) ([]protocol.CallHierarchyIncomingCall, error) {
	panic("unimplemented")
}

func (j *JavaLS) OutgoingCalls(ctx context.Context, params *protocol.CallHierarchyOutgoingCallsParams) ([]protocol.CallHierarchyOutgoingCall, error) {
	panic("unimplemented")
}

func (j *JavaLS) SemanticTokensFull(ctx context.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	panic("unimplemented")
}

func (j *JavaLS) SemanticTokensFullDelta(ctx context.Context, params *protocol.SemanticTokensDeltaParams) (interface{}, error) {
	panic("unimplemented")
}

func (j *JavaLS) SemanticTokensRange(ctx context.Context, params *protocol.SemanticTokensRangeParams) (*protocol.SemanticTokens, error) {
	panic("unimplemented")
}

func (j *JavaLS) SemanticTokensRefresh(ctx context.Context) error {
	panic("unimplemented")
}

func (j *JavaLS) LinkedEditingRange(ctx context.Context, params *protocol.LinkedEditingRangeParams) (*protocol.LinkedEditingRanges, error) {
	panic("unimplemented")
}

func (j *JavaLS) Moniker(ctx context.Context, params *protocol.MonikerParams) ([]protocol.Moniker, error) {
	panic("unimplemented")
}

func (j *JavaLS) Request(ctx context.Context, method string, params interface{}) (interface{}, error) {
	panic("unimplemented")
}
