// Code generated by MockGen. DO NOT EDIT.
// Source: types/invariant.go
//
// Generated by this command:
//
//	mockgen -source=types/invariant.go -package mock -destination testutil/mock/types_invariant.go
//

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	types "github.com/cosmos/cosmos-sdk/types"
	gomock "go.uber.org/mock/gomock"
)

// MockInvariantRegistry is a mock of InvariantRegistry interface.
type MockInvariantRegistry struct {
	ctrl     *gomock.Controller
	recorder *MockInvariantRegistryMockRecorder
	isgomock struct{}
}

// MockInvariantRegistryMockRecorder is the mock recorder for MockInvariantRegistry.
type MockInvariantRegistryMockRecorder struct {
	mock *MockInvariantRegistry
}

// NewMockInvariantRegistry creates a new mock instance.
func NewMockInvariantRegistry(ctrl *gomock.Controller) *MockInvariantRegistry {
	mock := &MockInvariantRegistry{ctrl: ctrl}
	mock.recorder = &MockInvariantRegistryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInvariantRegistry) EXPECT() *MockInvariantRegistryMockRecorder {
	return m.recorder
}

// RegisterRoute mocks base method.
func (m *MockInvariantRegistry) RegisterRoute(moduleName, route string, invar types.Invariant) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RegisterRoute", moduleName, route, invar)
}

// RegisterRoute indicates an expected call of RegisterRoute.
func (mr *MockInvariantRegistryMockRecorder) RegisterRoute(moduleName, route, invar any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterRoute", reflect.TypeOf((*MockInvariantRegistry)(nil).RegisterRoute), moduleName, route, invar)
}
