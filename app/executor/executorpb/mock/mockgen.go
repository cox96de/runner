// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cox96de/runner/app/executor/executorpb (interfaces: ExecutorClient)
//
// Generated by this command:
//
//	mockgen -destination mock/mockgen.go -package mock . ExecutorClient
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	executorpb "github.com/cox96de/runner/app/executor/executorpb"
	gomock "go.uber.org/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockExecutorClient is a mock of ExecutorClient interface.
type MockExecutorClient struct {
	ctrl     *gomock.Controller
	recorder *MockExecutorClientMockRecorder
}

// MockExecutorClientMockRecorder is the mock recorder for MockExecutorClient.
type MockExecutorClientMockRecorder struct {
	mock *MockExecutorClient
}

// NewMockExecutorClient creates a new mock instance.
func NewMockExecutorClient(ctrl *gomock.Controller) *MockExecutorClient {
	mock := &MockExecutorClient{ctrl: ctrl}
	mock.recorder = &MockExecutorClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExecutorClient) EXPECT() *MockExecutorClientMockRecorder {
	return m.recorder
}

// Environment mocks base method.
func (m *MockExecutorClient) Environment(arg0 context.Context, arg1 *executorpb.EnvironmentRequest, arg2 ...grpc.CallOption) (*executorpb.EnvironmentResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Environment", varargs...)
	ret0, _ := ret[0].(*executorpb.EnvironmentResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Environment indicates an expected call of Environment.
func (mr *MockExecutorClientMockRecorder) Environment(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Environment", reflect.TypeOf((*MockExecutorClient)(nil).Environment), varargs...)
}

// GetCommandLog mocks base method.
func (m *MockExecutorClient) GetCommandLog(arg0 context.Context, arg1 *executorpb.GetCommandLogRequest, arg2 ...grpc.CallOption) (executorpb.Executor_GetCommandLogClient, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetCommandLog", varargs...)
	ret0, _ := ret[0].(executorpb.Executor_GetCommandLogClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCommandLog indicates an expected call of GetCommandLog.
func (mr *MockExecutorClientMockRecorder) GetCommandLog(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCommandLog", reflect.TypeOf((*MockExecutorClient)(nil).GetCommandLog), varargs...)
}

// GetRuntimeInfo mocks base method.
func (m *MockExecutorClient) GetRuntimeInfo(arg0 context.Context, arg1 *executorpb.GetRuntimeInfoRequest, arg2 ...grpc.CallOption) (*executorpb.GetRuntimeInfoResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetRuntimeInfo", varargs...)
	ret0, _ := ret[0].(*executorpb.GetRuntimeInfoResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRuntimeInfo indicates an expected call of GetRuntimeInfo.
func (mr *MockExecutorClientMockRecorder) GetRuntimeInfo(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRuntimeInfo", reflect.TypeOf((*MockExecutorClient)(nil).GetRuntimeInfo), varargs...)
}

// Ping mocks base method.
func (m *MockExecutorClient) Ping(arg0 context.Context, arg1 *executorpb.PingRequest, arg2 ...grpc.CallOption) (*executorpb.PingResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Ping", varargs...)
	ret0, _ := ret[0].(*executorpb.PingResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Ping indicates an expected call of Ping.
func (mr *MockExecutorClientMockRecorder) Ping(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockExecutorClient)(nil).Ping), varargs...)
}

// StartCommand mocks base method.
func (m *MockExecutorClient) StartCommand(arg0 context.Context, arg1 *executorpb.StartCommandRequest, arg2 ...grpc.CallOption) (*executorpb.StartCommandResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "StartCommand", varargs...)
	ret0, _ := ret[0].(*executorpb.StartCommandResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StartCommand indicates an expected call of StartCommand.
func (mr *MockExecutorClientMockRecorder) StartCommand(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartCommand", reflect.TypeOf((*MockExecutorClient)(nil).StartCommand), varargs...)
}

// WaitCommand mocks base method.
func (m *MockExecutorClient) WaitCommand(arg0 context.Context, arg1 *executorpb.WaitCommandRequest, arg2 ...grpc.CallOption) (*executorpb.WaitCommandResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "WaitCommand", varargs...)
	ret0, _ := ret[0].(*executorpb.WaitCommandResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WaitCommand indicates an expected call of WaitCommand.
func (mr *MockExecutorClientMockRecorder) WaitCommand(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitCommand", reflect.TypeOf((*MockExecutorClient)(nil).WaitCommand), varargs...)
}
