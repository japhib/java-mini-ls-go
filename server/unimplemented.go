package server

import (
	"context"
	"go.lsp.dev/protocol"
)

// This file contains initialize/shutdown type calls

func (j *JavaLS) WorkDoneProgressCancel(ctx context.Context, params *protocol.WorkDoneProgressCancelParams) error {
	panic("WorkDoneProgressCancel unimplemented")
}

func (j *JavaLS) LogTrace(ctx context.Context, params *protocol.LogTraceParams) error {
	panic("LogTrace unimplemented")
}

func (j *JavaLS) SetTrace(ctx context.Context, params *protocol.SetTraceParams) error {
	panic("SetTrace unimplemented")
}

func (j *JavaLS) CodeAction(ctx context.Context, params *protocol.CodeActionParams) ([]protocol.CodeAction, error) {
	panic("CodeAction unimplemented")
}

func (j *JavaLS) CodeLens(ctx context.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	panic("CodeLens unimplemented")
}

func (j *JavaLS) CodeLensResolve(ctx context.Context, params *protocol.CodeLens) (*protocol.CodeLens, error) {
	panic("CodeLensResolve unimplemented")
}

func (j *JavaLS) ColorPresentation(ctx context.Context, params *protocol.ColorPresentationParams) ([]protocol.ColorPresentation, error) {
	panic("ColorPresentation unimplemented")
}

func (j *JavaLS) CompletionResolve(ctx context.Context, params *protocol.CompletionItem) (*protocol.CompletionItem, error) {
	panic("CompletionResolve unimplemented")
}

func (j *JavaLS) Declaration(ctx context.Context, params *protocol.DeclarationParams) ([]protocol.Location, error) {
	panic("Declaration unimplemented")
}

func (j *JavaLS) DidChangeConfiguration(ctx context.Context, params *protocol.DidChangeConfigurationParams) error {
	panic("DidChangeConfiguration unimplemented")
}

func (j *JavaLS) DidChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
	panic("DidChangeWatchedFiles unimplemented")
}

func (j *JavaLS) DidChangeWorkspaceFolders(ctx context.Context, params *protocol.DidChangeWorkspaceFoldersParams) error {
	j.log.Info("DidChangeWorkspaceFolders unimplemented")
	return nil
}

func (j *JavaLS) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
	j.log.Info("DidSave unimplemented")
	return nil
}

func (j *JavaLS) DocumentColor(ctx context.Context, params *protocol.DocumentColorParams) ([]protocol.ColorInformation, error) {
	panic("DocumentColor unimplemented")
}

func (j *JavaLS) DocumentHighlight(ctx context.Context, params *protocol.DocumentHighlightParams) ([]protocol.DocumentHighlight, error) {
	panic("DocumentHighlight unimplemented")
}

func (j *JavaLS) DocumentLink(ctx context.Context, params *protocol.DocumentLinkParams) ([]protocol.DocumentLink, error) {
	panic("DocumentLink unimplemented")
}

func (j *JavaLS) DocumentLinkResolve(ctx context.Context, params *protocol.DocumentLink) (*protocol.DocumentLink, error) {
	panic("DocumentLinkResolve unimplemented")
}

func (j *JavaLS) ExecuteCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (interface{}, error) {
	panic("ExecuteCommand unimplemented")
}

func (j *JavaLS) FoldingRanges(ctx context.Context, params *protocol.FoldingRangeParams) ([]protocol.FoldingRange, error) {
	panic("FoldingRanges unimplemented")
}

func (j *JavaLS) Formatting(ctx context.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	panic("Formatting unimplemented")
}

func (j *JavaLS) Implementation(ctx context.Context, params *protocol.ImplementationParams) ([]protocol.Location, error) {
	panic("Implementation unimplemented")
}

func (j *JavaLS) OnTypeFormatting(ctx context.Context, params *protocol.DocumentOnTypeFormattingParams) ([]protocol.TextEdit, error) {
	panic("OnTypeFormatting unimplemented")
}

func (j *JavaLS) PrepareRename(ctx context.Context, params *protocol.PrepareRenameParams) (*protocol.Range, error) {
	panic("PrepareRename unimplemented")
}

func (j *JavaLS) RangeFormatting(ctx context.Context, params *protocol.DocumentRangeFormattingParams) ([]protocol.TextEdit, error) {
	panic("RangeFormatting unimplemented")
}

func (j *JavaLS) Rename(ctx context.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	panic("Rename unimplemented")
}

func (j *JavaLS) SignatureHelp(ctx context.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	panic("SignatureHelp unimplemented")
}

func (j *JavaLS) Symbols(ctx context.Context, params *protocol.WorkspaceSymbolParams) ([]protocol.SymbolInformation, error) {
	panic("Symbols unimplemented")
}

func (j *JavaLS) TypeDefinition(ctx context.Context, params *protocol.TypeDefinitionParams) ([]protocol.Location, error) {
	panic("TypeDefinition unimplemented")
}

func (j *JavaLS) WillSave(ctx context.Context, params *protocol.WillSaveTextDocumentParams) error {
	panic("WillSave unimplemented")
}

func (j *JavaLS) WillSaveWaitUntil(ctx context.Context, params *protocol.WillSaveTextDocumentParams) ([]protocol.TextEdit, error) {
	panic("WillSaveWaitUntil unimplemented")
}

func (j *JavaLS) ShowDocument(ctx context.Context, params *protocol.ShowDocumentParams) (*protocol.ShowDocumentResult, error) {
	panic("ShowDocument unimplemented")
}

func (j *JavaLS) WillCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) (*protocol.WorkspaceEdit, error) {
	panic("WillCreateFiles unimplemented")
}

func (j *JavaLS) DidCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) error {
	panic("DidCreateFiles unimplemented")
}

func (j *JavaLS) WillRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) (*protocol.WorkspaceEdit, error) {
	panic("WillRenameFiles unimplemented")
}

func (j *JavaLS) DidRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) error {
	panic("DidRenameFiles unimplemented")
}

func (j *JavaLS) WillDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) (*protocol.WorkspaceEdit, error) {
	panic("WillDeleteFiles unimplemented")
}

func (j *JavaLS) DidDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) error {
	panic("DidDeleteFiles unimplemented")
}

func (j *JavaLS) CodeLensRefresh(ctx context.Context) error {
	panic("CodeLensRefresh unimplemented")
}

func (j *JavaLS) PrepareCallHierarchy(ctx context.Context, params *protocol.CallHierarchyPrepareParams) ([]protocol.CallHierarchyItem, error) {
	panic("PrepareCallHierarchy unimplemented")
}

func (j *JavaLS) IncomingCalls(ctx context.Context, params *protocol.CallHierarchyIncomingCallsParams) ([]protocol.CallHierarchyIncomingCall, error) {
	panic("IncomingCalls unimplemented")
}

func (j *JavaLS) OutgoingCalls(ctx context.Context, params *protocol.CallHierarchyOutgoingCallsParams) ([]protocol.CallHierarchyOutgoingCall, error) {
	panic("OutgoingCalls unimplemented")
}

func (j *JavaLS) SemanticTokensFull(ctx context.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	panic("SemanticTokensFull unimplemented")
}

func (j *JavaLS) SemanticTokensFullDelta(ctx context.Context, params *protocol.SemanticTokensDeltaParams) (interface{}, error) {
	panic("SemanticTokensFullDelta unimplemented")
}

func (j *JavaLS) SemanticTokensRange(ctx context.Context, params *protocol.SemanticTokensRangeParams) (*protocol.SemanticTokens, error) {
	panic("SemanticTokensRange unimplemented")
}

func (j *JavaLS) SemanticTokensRefresh(ctx context.Context) error {
	panic("SemanticTokensRefresh unimplemented")
}

func (j *JavaLS) LinkedEditingRange(ctx context.Context, params *protocol.LinkedEditingRangeParams) (*protocol.LinkedEditingRanges, error) {
	panic("LinkedEditingRange unimplemented")
}

func (j *JavaLS) Moniker(ctx context.Context, params *protocol.MonikerParams) ([]protocol.Moniker, error) {
	panic("Moniker unimplemented")
}

func (j *JavaLS) Request(ctx context.Context, method string, params interface{}) (interface{}, error) {
	panic("Request unimplemented")
}
