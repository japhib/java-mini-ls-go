// Code generated by MockGen. DO NOT EDIT.
// Source: ../diagnostics_publisher.go

// Package mock_server is a generated GoMock package.
package server

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	protocol "go.lsp.dev/protocol"
)

// MockDiagnosticsPublisher is a mock of DiagnosticsPublisher interface.
type MockDiagnosticsPublisher struct {
	ctrl     *gomock.Controller
	recorder *MockDiagnosticsPublisherMockRecorder
}

// MockDiagnosticsPublisherMockRecorder is the mock recorder for MockDiagnosticsPublisher.
type MockDiagnosticsPublisherMockRecorder struct {
	mock *MockDiagnosticsPublisher
}

// NewMockDiagnosticsPublisher creates a new mock instance.
func NewMockDiagnosticsPublisher(ctrl *gomock.Controller) *MockDiagnosticsPublisher {
	mock := &MockDiagnosticsPublisher{ctrl: ctrl}
	mock.recorder = &MockDiagnosticsPublisherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDiagnosticsPublisher) EXPECT() *MockDiagnosticsPublisherMockRecorder {
	return m.recorder
}

// PublishDiagnostics mocks base method.
func (m *MockDiagnosticsPublisher) PublishDiagnostics(j *JavaLS, textDocument protocol.TextDocumentItem, diagnostics []protocol.Diagnostic) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PublishDiagnostics", j, textDocument, diagnostics)
}

// PublishDiagnostics indicates an expected call of PublishDiagnostics.
func (mr *MockDiagnosticsPublisherMockRecorder) PublishDiagnostics(j, textDocument, diagnostics interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PublishDiagnostics", reflect.TypeOf((*MockDiagnosticsPublisher)(nil).PublishDiagnostics), j, textDocument, diagnostics)
}
