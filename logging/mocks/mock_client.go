// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/kitagry/gcp-telemetry-mcp/logging (interfaces: LoggingClient,LoggingClientInterface)
//
// Generated by this command:
//
//	mockgen -destination=mocks/mock_client.go -package=mocks github.com/kitagry/gcp-telemetry-mcp/logging LoggingClient,LoggingClientInterface
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	logging "github.com/kitagry/gcp-telemetry-mcp/logging"
	gomock "go.uber.org/mock/gomock"
)

// MockLoggingClient is a mock of LoggingClient interface.
type MockLoggingClient struct {
	ctrl     *gomock.Controller
	recorder *MockLoggingClientMockRecorder
	isgomock struct{}
}

// MockLoggingClientMockRecorder is the mock recorder for MockLoggingClient.
type MockLoggingClientMockRecorder struct {
	mock *MockLoggingClient
}

// NewMockLoggingClient creates a new mock instance.
func NewMockLoggingClient(ctrl *gomock.Controller) *MockLoggingClient {
	mock := &MockLoggingClient{ctrl: ctrl}
	mock.recorder = &MockLoggingClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLoggingClient) EXPECT() *MockLoggingClientMockRecorder {
	return m.recorder
}

// ListEntries mocks base method.
func (m *MockLoggingClient) ListEntries(ctx context.Context, req logging.ListEntriesRequest) ([]logging.LogEntry, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListEntries", ctx, req)
	ret0, _ := ret[0].([]logging.LogEntry)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEntries indicates an expected call of ListEntries.
func (mr *MockLoggingClientMockRecorder) ListEntries(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEntries", reflect.TypeOf((*MockLoggingClient)(nil).ListEntries), ctx, req)
}

// WriteEntry mocks base method.
func (m *MockLoggingClient) WriteEntry(ctx context.Context, logName string, entry logging.LogEntry) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteEntry", ctx, logName, entry)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteEntry indicates an expected call of WriteEntry.
func (mr *MockLoggingClientMockRecorder) WriteEntry(ctx, logName, entry any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteEntry", reflect.TypeOf((*MockLoggingClient)(nil).WriteEntry), ctx, logName, entry)
}

// MockLoggingClientInterface is a mock of LoggingClientInterface interface.
type MockLoggingClientInterface struct {
	ctrl     *gomock.Controller
	recorder *MockLoggingClientInterfaceMockRecorder
	isgomock struct{}
}

// MockLoggingClientInterfaceMockRecorder is the mock recorder for MockLoggingClientInterface.
type MockLoggingClientInterfaceMockRecorder struct {
	mock *MockLoggingClientInterface
}

// NewMockLoggingClientInterface creates a new mock instance.
func NewMockLoggingClientInterface(ctrl *gomock.Controller) *MockLoggingClientInterface {
	mock := &MockLoggingClientInterface{ctrl: ctrl}
	mock.recorder = &MockLoggingClientInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLoggingClientInterface) EXPECT() *MockLoggingClientInterfaceMockRecorder {
	return m.recorder
}

// ListEntries mocks base method.
func (m *MockLoggingClientInterface) ListEntries(ctx context.Context, req logging.ListEntriesRequest) ([]logging.LogEntry, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListEntries", ctx, req)
	ret0, _ := ret[0].([]logging.LogEntry)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEntries indicates an expected call of ListEntries.
func (mr *MockLoggingClientInterfaceMockRecorder) ListEntries(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEntries", reflect.TypeOf((*MockLoggingClientInterface)(nil).ListEntries), ctx, req)
}

// WriteEntry mocks base method.
func (m *MockLoggingClientInterface) WriteEntry(ctx context.Context, logName string, entry logging.LogEntry) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteEntry", ctx, logName, entry)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteEntry indicates an expected call of WriteEntry.
func (mr *MockLoggingClientInterfaceMockRecorder) WriteEntry(ctx, logName, entry any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteEntry", reflect.TypeOf((*MockLoggingClientInterface)(nil).WriteEntry), ctx, logName, entry)
}
