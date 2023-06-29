// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cox96de/runner/testtool (interfaces: TestingT)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockTestingT is a mock of TestingT interface.
type MockTestingT struct {
	ctrl     *gomock.Controller
	recorder *MockTestingTMockRecorder
}

// MockTestingTMockRecorder is the mock recorder for MockTestingT.
type MockTestingTMockRecorder struct {
	mock *MockTestingT
}

// NewMockTestingT creates a new mock instance.
func NewMockTestingT(ctrl *gomock.Controller) *MockTestingT {
	mock := &MockTestingT{ctrl: ctrl}
	mock.recorder = &MockTestingTMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTestingT) EXPECT() *MockTestingTMockRecorder {
	return m.recorder
}

// Fail mocks base method.
func (m *MockTestingT) Fail() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Fail")
}

// Fail indicates an expected call of Fail.
func (mr *MockTestingTMockRecorder) Fail() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fail", reflect.TypeOf((*MockTestingT)(nil).Fail))
}

// FailNow mocks base method.
func (m *MockTestingT) FailNow() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "FailNow")
}

// FailNow indicates an expected call of FailNow.
func (mr *MockTestingTMockRecorder) FailNow() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FailNow", reflect.TypeOf((*MockTestingT)(nil).FailNow))
}

// Helper mocks base method.
func (m *MockTestingT) Helper() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Helper")
}

// Helper indicates an expected call of Helper.
func (mr *MockTestingTMockRecorder) Helper() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Helper", reflect.TypeOf((*MockTestingT)(nil).Helper))
}

// Log mocks base method.
func (m *MockTestingT) Log(arg0 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Log", varargs...)
}

// Log indicates an expected call of Log.
func (mr *MockTestingTMockRecorder) Log(arg0 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Log", reflect.TypeOf((*MockTestingT)(nil).Log), arg0...)
}
