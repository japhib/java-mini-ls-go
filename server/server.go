package server

import (
	"context"
	"go.lsp.dev/protocol"
)

type JavaLS struct{}

func (j *JavaLS) Initialize(ctx context.Context, params *protocol.InitializeParams) (result *protocol.InitializeResult, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) Initialized(ctx context.Context, params *protocol.InitializedParams) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) Shutdown(ctx context.Context) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) Exit(ctx context.Context) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) WorkDoneProgressCancel(ctx context.Context, params *protocol.WorkDoneProgressCancelParams) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) LogTrace(ctx context.Context, params *protocol.LogTraceParams) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) SetTrace(ctx context.Context, params *protocol.SetTraceParams) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) CodeAction(ctx context.Context, params *protocol.CodeActionParams) (result []protocol.CodeAction, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) CodeLens(ctx context.Context, params *protocol.CodeLensParams) (result []protocol.CodeLens, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) CodeLensResolve(ctx context.Context, params *protocol.CodeLens) (result *protocol.CodeLens, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) ColorPresentation(ctx context.Context, params *protocol.ColorPresentationParams) (result []protocol.ColorPresentation, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) Completion(ctx context.Context, params *protocol.CompletionParams) (result *protocol.CompletionList, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) CompletionResolve(ctx context.Context, params *protocol.CompletionItem) (result *protocol.CompletionItem, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) Declaration(ctx context.Context, params *protocol.DeclarationParams) (result []protocol.Location, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) Definition(ctx context.Context, params *protocol.DefinitionParams) (result []protocol.Location, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) DidChangeConfiguration(ctx context.Context, params *protocol.DidChangeConfigurationParams) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) DidChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) DidChangeWorkspaceFolders(ctx context.Context, params *protocol.DidChangeWorkspaceFoldersParams) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) DocumentColor(ctx context.Context, params *protocol.DocumentColorParams) (result []protocol.ColorInformation, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) DocumentHighlight(ctx context.Context, params *protocol.DocumentHighlightParams) (result []protocol.DocumentHighlight, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) DocumentLink(ctx context.Context, params *protocol.DocumentLinkParams) (result []protocol.DocumentLink, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) DocumentLinkResolve(ctx context.Context, params *protocol.DocumentLink) (result *protocol.DocumentLink, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) DocumentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) (result []interface{}, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) ExecuteCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (result interface{}, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) FoldingRanges(ctx context.Context, params *protocol.FoldingRangeParams) (result []protocol.FoldingRange, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) Formatting(ctx context.Context, params *protocol.DocumentFormattingParams) (result []protocol.TextEdit, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) Hover(ctx context.Context, params *protocol.HoverParams) (result *protocol.Hover, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) Implementation(ctx context.Context, params *protocol.ImplementationParams) (result []protocol.Location, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) OnTypeFormatting(ctx context.Context, params *protocol.DocumentOnTypeFormattingParams) (result []protocol.TextEdit, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) PrepareRename(ctx context.Context, params *protocol.PrepareRenameParams) (result *protocol.Range, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) RangeFormatting(ctx context.Context, params *protocol.DocumentRangeFormattingParams) (result []protocol.TextEdit, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) References(ctx context.Context, params *protocol.ReferenceParams) (result []protocol.Location, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) Rename(ctx context.Context, params *protocol.RenameParams) (result *protocol.WorkspaceEdit, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) SignatureHelp(ctx context.Context, params *protocol.SignatureHelpParams) (result *protocol.SignatureHelp, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) Symbols(ctx context.Context, params *protocol.WorkspaceSymbolParams) (result []protocol.SymbolInformation, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) TypeDefinition(ctx context.Context, params *protocol.TypeDefinitionParams) (result []protocol.Location, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) WillSave(ctx context.Context, params *protocol.WillSaveTextDocumentParams) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) WillSaveWaitUntil(ctx context.Context, params *protocol.WillSaveTextDocumentParams) (result []protocol.TextEdit, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) ShowDocument(ctx context.Context, params *protocol.ShowDocumentParams) (result *protocol.ShowDocumentResult, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) WillCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) (result *protocol.WorkspaceEdit, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) DidCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) WillRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) (result *protocol.WorkspaceEdit, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) DidRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) WillDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) (result *protocol.WorkspaceEdit, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) DidDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) CodeLensRefresh(ctx context.Context) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) PrepareCallHierarchy(ctx context.Context, params *protocol.CallHierarchyPrepareParams) (result []protocol.CallHierarchyItem, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) IncomingCalls(ctx context.Context, params *protocol.CallHierarchyIncomingCallsParams) (result []protocol.CallHierarchyIncomingCall, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) OutgoingCalls(ctx context.Context, params *protocol.CallHierarchyOutgoingCallsParams) (result []protocol.CallHierarchyOutgoingCall, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) SemanticTokensFull(ctx context.Context, params *protocol.SemanticTokensParams) (result *protocol.SemanticTokens, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) SemanticTokensFullDelta(ctx context.Context, params *protocol.SemanticTokensDeltaParams) (result interface{}, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) SemanticTokensRange(ctx context.Context, params *protocol.SemanticTokensRangeParams) (result *protocol.SemanticTokens, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) SemanticTokensRefresh(ctx context.Context) (err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) LinkedEditingRange(ctx context.Context, params *protocol.LinkedEditingRangeParams) (result *protocol.LinkedEditingRanges, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) Moniker(ctx context.Context, params *protocol.MonikerParams) (result []protocol.Moniker, err error) {
	//TODO implement me
	panic("implement me")
}

func (j *JavaLS) Request(ctx context.Context, method string, params interface{}) (result interface{}, err error) {
	//TODO implement me
	panic("implement me")
}
