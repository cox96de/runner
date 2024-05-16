// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cox96de/runner/api (interfaces: ServerClient)
//
// Generated by this command:
//
//	mockgen -destination mock/server_mockgen.go -typed -package mock . ServerClient
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	api "github.com/cox96de/runner/api"
	gomock "go.uber.org/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockServerClient is a mock of ServerClient interface.
type MockServerClient struct {
	ctrl     *gomock.Controller
	recorder *MockServerClientMockRecorder
}

// MockServerClientMockRecorder is the mock recorder for MockServerClient.
type MockServerClientMockRecorder struct {
	mock *MockServerClient
}

// NewMockServerClient creates a new mock instance.
func NewMockServerClient(ctrl *gomock.Controller) *MockServerClient {
	mock := &MockServerClient{ctrl: ctrl}
	mock.recorder = &MockServerClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockServerClient) EXPECT() *MockServerClientMockRecorder {
	return m.recorder
}

// CreatePipeline mocks base method.
func (m *MockServerClient) CreatePipeline(arg0 context.Context, arg1 *api.CreatePipelineRequest, arg2 ...grpc.CallOption) (*api.CreatePipelineResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreatePipeline", varargs...)
	ret0, _ := ret[0].(*api.CreatePipelineResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreatePipeline indicates an expected call of CreatePipeline.
func (mr *MockServerClientMockRecorder) CreatePipeline(arg0, arg1 any, arg2 ...any) *MockServerClientCreatePipelineCall {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePipeline", reflect.TypeOf((*MockServerClient)(nil).CreatePipeline), varargs...)
	return &MockServerClientCreatePipelineCall{Call: call}
}

// MockServerClientCreatePipelineCall wrap *gomock.Call
type MockServerClientCreatePipelineCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServerClientCreatePipelineCall) Return(arg0 *api.CreatePipelineResponse, arg1 error) *MockServerClientCreatePipelineCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServerClientCreatePipelineCall) Do(f func(context.Context, *api.CreatePipelineRequest, ...grpc.CallOption) (*api.CreatePipelineResponse, error)) *MockServerClientCreatePipelineCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServerClientCreatePipelineCall) DoAndReturn(f func(context.Context, *api.CreatePipelineRequest, ...grpc.CallOption) (*api.CreatePipelineResponse, error)) *MockServerClientCreatePipelineCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetLogLines mocks base method.
func (m *MockServerClient) GetLogLines(arg0 context.Context, arg1 *api.GetLogLinesRequest, arg2 ...grpc.CallOption) (*api.GetLogLinesResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetLogLines", varargs...)
	ret0, _ := ret[0].(*api.GetLogLinesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLogLines indicates an expected call of GetLogLines.
func (mr *MockServerClientMockRecorder) GetLogLines(arg0, arg1 any, arg2 ...any) *MockServerClientGetLogLinesCall {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLogLines", reflect.TypeOf((*MockServerClient)(nil).GetLogLines), varargs...)
	return &MockServerClientGetLogLinesCall{Call: call}
}

// MockServerClientGetLogLinesCall wrap *gomock.Call
type MockServerClientGetLogLinesCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServerClientGetLogLinesCall) Return(arg0 *api.GetLogLinesResponse, arg1 error) *MockServerClientGetLogLinesCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServerClientGetLogLinesCall) Do(f func(context.Context, *api.GetLogLinesRequest, ...grpc.CallOption) (*api.GetLogLinesResponse, error)) *MockServerClientGetLogLinesCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServerClientGetLogLinesCall) DoAndReturn(f func(context.Context, *api.GetLogLinesRequest, ...grpc.CallOption) (*api.GetLogLinesResponse, error)) *MockServerClientGetLogLinesCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ListJobExecutions mocks base method.
func (m *MockServerClient) ListJobExecutions(arg0 context.Context, arg1 *api.ListJobExecutionsRequest, arg2 ...grpc.CallOption) (*api.ListJobExecutionsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListJobExecutions", varargs...)
	ret0, _ := ret[0].(*api.ListJobExecutionsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListJobExecutions indicates an expected call of ListJobExecutions.
func (mr *MockServerClientMockRecorder) ListJobExecutions(arg0, arg1 any, arg2 ...any) *MockServerClientListJobExecutionsCall {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListJobExecutions", reflect.TypeOf((*MockServerClient)(nil).ListJobExecutions), varargs...)
	return &MockServerClientListJobExecutionsCall{Call: call}
}

// MockServerClientListJobExecutionsCall wrap *gomock.Call
type MockServerClientListJobExecutionsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServerClientListJobExecutionsCall) Return(arg0 *api.ListJobExecutionsResponse, arg1 error) *MockServerClientListJobExecutionsCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServerClientListJobExecutionsCall) Do(f func(context.Context, *api.ListJobExecutionsRequest, ...grpc.CallOption) (*api.ListJobExecutionsResponse, error)) *MockServerClientListJobExecutionsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServerClientListJobExecutionsCall) DoAndReturn(f func(context.Context, *api.ListJobExecutionsRequest, ...grpc.CallOption) (*api.ListJobExecutionsResponse, error)) *MockServerClientListJobExecutionsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// RequestJob mocks base method.
func (m *MockServerClient) RequestJob(arg0 context.Context, arg1 *api.RequestJobRequest, arg2 ...grpc.CallOption) (*api.RequestJobResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RequestJob", varargs...)
	ret0, _ := ret[0].(*api.RequestJobResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RequestJob indicates an expected call of RequestJob.
func (mr *MockServerClientMockRecorder) RequestJob(arg0, arg1 any, arg2 ...any) *MockServerClientRequestJobCall {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestJob", reflect.TypeOf((*MockServerClient)(nil).RequestJob), varargs...)
	return &MockServerClientRequestJobCall{Call: call}
}

// MockServerClientRequestJobCall wrap *gomock.Call
type MockServerClientRequestJobCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServerClientRequestJobCall) Return(arg0 *api.RequestJobResponse, arg1 error) *MockServerClientRequestJobCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServerClientRequestJobCall) Do(f func(context.Context, *api.RequestJobRequest, ...grpc.CallOption) (*api.RequestJobResponse, error)) *MockServerClientRequestJobCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServerClientRequestJobCall) DoAndReturn(f func(context.Context, *api.RequestJobRequest, ...grpc.CallOption) (*api.RequestJobResponse, error)) *MockServerClientRequestJobCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// UpdateJobExecution mocks base method.
func (m *MockServerClient) UpdateJobExecution(arg0 context.Context, arg1 *api.UpdateJobExecutionRequest, arg2 ...grpc.CallOption) (*api.UpdateJobExecutionResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateJobExecution", varargs...)
	ret0, _ := ret[0].(*api.UpdateJobExecutionResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateJobExecution indicates an expected call of UpdateJobExecution.
func (mr *MockServerClientMockRecorder) UpdateJobExecution(arg0, arg1 any, arg2 ...any) *MockServerClientUpdateJobExecutionCall {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateJobExecution", reflect.TypeOf((*MockServerClient)(nil).UpdateJobExecution), varargs...)
	return &MockServerClientUpdateJobExecutionCall{Call: call}
}

// MockServerClientUpdateJobExecutionCall wrap *gomock.Call
type MockServerClientUpdateJobExecutionCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServerClientUpdateJobExecutionCall) Return(arg0 *api.UpdateJobExecutionResponse, arg1 error) *MockServerClientUpdateJobExecutionCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServerClientUpdateJobExecutionCall) Do(f func(context.Context, *api.UpdateJobExecutionRequest, ...grpc.CallOption) (*api.UpdateJobExecutionResponse, error)) *MockServerClientUpdateJobExecutionCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServerClientUpdateJobExecutionCall) DoAndReturn(f func(context.Context, *api.UpdateJobExecutionRequest, ...grpc.CallOption) (*api.UpdateJobExecutionResponse, error)) *MockServerClientUpdateJobExecutionCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// UpdateStepExecution mocks base method.
func (m *MockServerClient) UpdateStepExecution(arg0 context.Context, arg1 *api.UpdateStepExecutionRequest, arg2 ...grpc.CallOption) (*api.UpdateStepExecutionResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateStepExecution", varargs...)
	ret0, _ := ret[0].(*api.UpdateStepExecutionResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateStepExecution indicates an expected call of UpdateStepExecution.
func (mr *MockServerClientMockRecorder) UpdateStepExecution(arg0, arg1 any, arg2 ...any) *MockServerClientUpdateStepExecutionCall {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateStepExecution", reflect.TypeOf((*MockServerClient)(nil).UpdateStepExecution), varargs...)
	return &MockServerClientUpdateStepExecutionCall{Call: call}
}

// MockServerClientUpdateStepExecutionCall wrap *gomock.Call
type MockServerClientUpdateStepExecutionCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServerClientUpdateStepExecutionCall) Return(arg0 *api.UpdateStepExecutionResponse, arg1 error) *MockServerClientUpdateStepExecutionCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServerClientUpdateStepExecutionCall) Do(f func(context.Context, *api.UpdateStepExecutionRequest, ...grpc.CallOption) (*api.UpdateStepExecutionResponse, error)) *MockServerClientUpdateStepExecutionCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServerClientUpdateStepExecutionCall) DoAndReturn(f func(context.Context, *api.UpdateStepExecutionRequest, ...grpc.CallOption) (*api.UpdateStepExecutionResponse, error)) *MockServerClientUpdateStepExecutionCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// UploadLogLines mocks base method.
func (m *MockServerClient) UploadLogLines(arg0 context.Context, arg1 *api.UpdateLogLinesRequest, arg2 ...grpc.CallOption) (*api.UpdateLogLinesResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UploadLogLines", varargs...)
	ret0, _ := ret[0].(*api.UpdateLogLinesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UploadLogLines indicates an expected call of UploadLogLines.
func (mr *MockServerClientMockRecorder) UploadLogLines(arg0, arg1 any, arg2 ...any) *MockServerClientUploadLogLinesCall {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadLogLines", reflect.TypeOf((*MockServerClient)(nil).UploadLogLines), varargs...)
	return &MockServerClientUploadLogLinesCall{Call: call}
}

// MockServerClientUploadLogLinesCall wrap *gomock.Call
type MockServerClientUploadLogLinesCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServerClientUploadLogLinesCall) Return(arg0 *api.UpdateLogLinesResponse, arg1 error) *MockServerClientUploadLogLinesCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServerClientUploadLogLinesCall) Do(f func(context.Context, *api.UpdateLogLinesRequest, ...grpc.CallOption) (*api.UpdateLogLinesResponse, error)) *MockServerClientUploadLogLinesCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServerClientUploadLogLinesCall) DoAndReturn(f func(context.Context, *api.UpdateLogLinesRequest, ...grpc.CallOption) (*api.UpdateLogLinesResponse, error)) *MockServerClientUploadLogLinesCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
