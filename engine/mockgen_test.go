// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cox96de/runner/engine (interfaces: Engine,Runner,Executor)

// Package engine is a generated GoMock package.
package engine

import (
	context "context"
	io "io"
	reflect "reflect"

	executor "github.com/cox96de/runner/internal/executor"
	model "github.com/cox96de/runner/internal/model"
	gomock "go.uber.org/mock/gomock"
)

// MockEngine is a mock of Engine interface.
type MockEngine struct {
	ctrl     *gomock.Controller
	recorder *MockEngineMockRecorder
}

// MockEngineMockRecorder is the mock recorder for MockEngine.
type MockEngineMockRecorder struct {
	mock *MockEngine
}

// NewMockEngine creates a new mock instance.
func NewMockEngine(ctrl *gomock.Controller) *MockEngine {
	mock := &MockEngine{ctrl: ctrl}
	mock.recorder = &MockEngineMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEngine) EXPECT() *MockEngineMockRecorder {
	return m.recorder
}

// CreateRunner mocks base method.
func (m *MockEngine) CreateRunner(arg0 context.Context, arg1 *RunnerSpec) (Runner, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateRunner", arg0, arg1)
	ret0, _ := ret[0].(Runner)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateRunner indicates an expected call of CreateRunner.
func (mr *MockEngineMockRecorder) CreateRunner(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateRunner", reflect.TypeOf((*MockEngine)(nil).CreateRunner), arg0, arg1)
}

// Ping mocks base method.
func (m *MockEngine) Ping(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockEngineMockRecorder) Ping(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockEngine)(nil).Ping), arg0)
}

// MockRunner is a mock of Runner interface.
type MockRunner struct {
	ctrl     *gomock.Controller
	recorder *MockRunnerMockRecorder
}

// MockRunnerMockRecorder is the mock recorder for MockRunner.
type MockRunnerMockRecorder struct {
	mock *MockRunner
}

// NewMockRunner creates a new mock instance.
func NewMockRunner(ctrl *gomock.Controller) *MockRunner {
	mock := &MockRunner{ctrl: ctrl}
	mock.recorder = &MockRunnerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRunner) EXPECT() *MockRunnerMockRecorder {
	return m.recorder
}

// GetExecutor mocks base method.
func (m *MockRunner) GetExecutor(arg0 context.Context, arg1 string) (Executor, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExecutor", arg0, arg1)
	ret0, _ := ret[0].(Executor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExecutor indicates an expected call of GetExecutor.
func (mr *MockRunnerMockRecorder) GetExecutor(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExecutor", reflect.TypeOf((*MockRunner)(nil).GetExecutor), arg0, arg1)
}

// Start mocks base method.
func (m *MockRunner) Start(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockRunnerMockRecorder) Start(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockRunner)(nil).Start), arg0)
}

// Stop mocks base method.
func (m *MockRunner) Stop(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockRunnerMockRecorder) Stop(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockRunner)(nil).Stop), arg0)
}

// MockExecutor is a mock of Executor interface.
type MockExecutor struct {
	ctrl     *gomock.Controller
	recorder *MockExecutorMockRecorder
}

// MockExecutorMockRecorder is the mock recorder for MockExecutor.
type MockExecutorMockRecorder struct {
	mock *MockExecutor
}

// NewMockExecutor creates a new mock instance.
func NewMockExecutor(ctrl *gomock.Controller) *MockExecutor {
	mock := &MockExecutor{ctrl: ctrl}
	mock.recorder = &MockExecutorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExecutor) EXPECT() *MockExecutorMockRecorder {
	return m.recorder
}

// GetCommandLogs mocks base method.
func (m *MockExecutor) GetCommandLogs(arg0 context.Context, arg1 string) io.ReadCloser {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCommandLogs", arg0, arg1)
	ret0, _ := ret[0].(io.ReadCloser)
	return ret0
}

// GetCommandLogs indicates an expected call of GetCommandLogs.
func (mr *MockExecutorMockRecorder) GetCommandLogs(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCommandLogs", reflect.TypeOf((*MockExecutor)(nil).GetCommandLogs), arg0, arg1)
}

// GetCommandStatus mocks base method.
func (m *MockExecutor) GetCommandStatus(arg0 context.Context, arg1 string) (*model.GetCommandStatusResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCommandStatus", arg0, arg1)
	ret0, _ := ret[0].(*model.GetCommandStatusResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCommandStatus indicates an expected call of GetCommandStatus.
func (mr *MockExecutorMockRecorder) GetCommandStatus(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCommandStatus", reflect.TypeOf((*MockExecutor)(nil).GetCommandStatus), arg0, arg1)
}

// Ping mocks base method.
func (m *MockExecutor) Ping(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockExecutorMockRecorder) Ping(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockExecutor)(nil).Ping), arg0)
}

// StartCommand mocks base method.
func (m *MockExecutor) StartCommand(arg0 context.Context, arg1 string, arg2 *executor.StartCommandRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartCommand", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// StartCommand indicates an expected call of StartCommand.
func (mr *MockExecutorMockRecorder) StartCommand(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartCommand", reflect.TypeOf((*MockExecutor)(nil).StartCommand), arg0, arg1, arg2)
}
