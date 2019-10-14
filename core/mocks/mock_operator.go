// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/v-braun/go2p/core (interfaces: Operator)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	core "github.com/v-braun/go2p/core"
	reflect "reflect"
)

// MockOperator is a mock of Operator interface
type MockOperator struct {
	ctrl     *gomock.Controller
	recorder *MockOperatorMockRecorder
}

// MockOperatorMockRecorder is the mock recorder for MockOperator
type MockOperatorMockRecorder struct {
	mock *MockOperator
}

// NewMockOperator creates a new mock instance
func NewMockOperator(ctrl *gomock.Controller) *MockOperator {
	mock := &MockOperator{ctrl: ctrl}
	mock.recorder = &MockOperatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockOperator) EXPECT() *MockOperatorMockRecorder {
	return m.recorder
}

// Dial mocks base method
func (m *MockOperator) Dial(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dial", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Dial indicates an expected call of Dial
func (mr *MockOperatorMockRecorder) Dial(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dial", reflect.TypeOf((*MockOperator)(nil).Dial), arg0, arg1)
}

// Install mocks base method
func (m *MockOperator) Install(arg0 *core.Network) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Install", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Install indicates an expected call of Install
func (mr *MockOperatorMockRecorder) Install(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Install", reflect.TypeOf((*MockOperator)(nil).Install), arg0)
}

// OnPeer mocks base method
func (m *MockOperator) OnPeer(arg0 func(core.Conn)) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "OnPeer", arg0)
}

// OnPeer indicates an expected call of OnPeer
func (mr *MockOperatorMockRecorder) OnPeer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnPeer", reflect.TypeOf((*MockOperator)(nil).OnPeer), arg0)
}

// Uninstall mocks base method
func (m *MockOperator) Uninstall() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Uninstall")
}

// Uninstall indicates an expected call of Uninstall
func (mr *MockOperatorMockRecorder) Uninstall() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Uninstall", reflect.TypeOf((*MockOperator)(nil).Uninstall))
}