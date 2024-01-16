// Code generated by MockGen. DO NOT EDIT.
// Source: x/auth/ante/expected_keepers.go

// Package testutil is a generated GoMock package.
package testutil

import (
	context "context"
	reflect "reflect"

	address "cosmossdk.io/core/address"
	types "github.com/cosmos/cosmos-sdk/types"
	types0 "github.com/cosmos/cosmos-sdk/x/auth/types"
	gomock "github.com/golang/mock/gomock"
)

// MockAccountKeeper is a mock of AccountKeeper interface.
type MockAccountKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockAccountKeeperMockRecorder
}

// MockAccountKeeperMockRecorder is the mock recorder for MockAccountKeeper.
type MockAccountKeeperMockRecorder struct {
	mock *MockAccountKeeper
}

// NewMockAccountKeeper creates a new mock instance.
func NewMockAccountKeeper(ctrl *gomock.Controller) *MockAccountKeeper {
	mock := &MockAccountKeeper{ctrl: ctrl}
	mock.recorder = &MockAccountKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccountKeeper) EXPECT() *MockAccountKeeperMockRecorder {
	return m.recorder
}

// AddressCodec mocks base method.
func (m *MockAccountKeeper) AddressCodec() address.Codec {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddressCodec")
	ret0, _ := ret[0].(address.Codec)
	return ret0
}

// AddressCodec indicates an expected call of AddressCodec.
func (mr *MockAccountKeeperMockRecorder) AddressCodec() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddressCodec", reflect.TypeOf((*MockAccountKeeper)(nil).AddressCodec))
}

// GetAccount mocks base method.
func (m *MockAccountKeeper) GetAccount(ctx context.Context, addr types.AccAddress) types.AccountI {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccount", ctx, addr)
	ret0, _ := ret[0].(types.AccountI)
	return ret0
}

// GetAccount indicates an expected call of GetAccount.
func (mr *MockAccountKeeperMockRecorder) GetAccount(ctx, addr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccount", reflect.TypeOf((*MockAccountKeeper)(nil).GetAccount), ctx, addr)
}

// GetModuleAddress mocks base method.
func (m *MockAccountKeeper) GetModuleAddress(moduleName string) types.AccAddress {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetModuleAddress", moduleName)
	ret0, _ := ret[0].(types.AccAddress)
	return ret0
}

// GetModuleAddress indicates an expected call of GetModuleAddress.
func (mr *MockAccountKeeperMockRecorder) GetModuleAddress(moduleName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetModuleAddress", reflect.TypeOf((*MockAccountKeeper)(nil).GetModuleAddress), moduleName)
}

// GetParams mocks base method.
func (m *MockAccountKeeper) GetParams(ctx context.Context) types0.Params {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetParams", ctx)
	ret0, _ := ret[0].(types0.Params)
	return ret0
}

// GetParams indicates an expected call of GetParams.
func (mr *MockAccountKeeperMockRecorder) GetParams(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetParams", reflect.TypeOf((*MockAccountKeeper)(nil).GetParams), ctx)
}

// SetAccount mocks base method.
func (m *MockAccountKeeper) SetAccount(ctx context.Context, acc types.AccountI) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetAccount", ctx, acc)
}

// SetAccount indicates an expected call of SetAccount.
func (mr *MockAccountKeeperMockRecorder) SetAccount(ctx, acc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAccount", reflect.TypeOf((*MockAccountKeeper)(nil).SetAccount), ctx, acc)
}

// MockFeegrantKeeper is a mock of FeegrantKeeper interface.
type MockFeegrantKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockFeegrantKeeperMockRecorder
}

// MockFeegrantKeeperMockRecorder is the mock recorder for MockFeegrantKeeper.
type MockFeegrantKeeperMockRecorder struct {
	mock *MockFeegrantKeeper
}

// NewMockFeegrantKeeper creates a new mock instance.
func NewMockFeegrantKeeper(ctrl *gomock.Controller) *MockFeegrantKeeper {
	mock := &MockFeegrantKeeper{ctrl: ctrl}
	mock.recorder = &MockFeegrantKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFeegrantKeeper) EXPECT() *MockFeegrantKeeperMockRecorder {
	return m.recorder
}

// UseGrantedFees mocks base method.
func (m *MockFeegrantKeeper) UseGrantedFees(ctx context.Context, granter, grantee types.AccAddress, fee types.Coins, msgs []types.Msg) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UseGrantedFees", ctx, granter, grantee, fee, msgs)
	ret0, _ := ret[0].(error)
	return ret0
}

// UseGrantedFees indicates an expected call of UseGrantedFees.
func (mr *MockFeegrantKeeperMockRecorder) UseGrantedFees(ctx, granter, grantee, fee, msgs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UseGrantedFees", reflect.TypeOf((*MockFeegrantKeeper)(nil).UseGrantedFees), ctx, granter, grantee, fee, msgs)
}

// MockFeeCollector is a mock of FeeCollector interface.
type MockFeeCollector struct {
	ctrl     *gomock.Controller
	recorder *MockFeeCollectorMockRecorder
}

// MockFeeCollectorMockRecorder is the mock recorder for MockFeeCollector.
type MockFeeCollectorMockRecorder struct {
	mock *MockFeeCollector
}

// NewMockFeeCollector creates a new mock instance.
func NewMockFeeCollector(ctrl *gomock.Controller) *MockFeeCollector {
	mock := &MockFeeCollector{ctrl: ctrl}
	mock.recorder = &MockFeeCollectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFeeCollector) EXPECT() *MockFeeCollectorMockRecorder {
	return m.recorder
}

// GetModuleAddress mocks base method.
func (m *MockFeeCollector) GetModuleAddress(arg0 string) types.AccAddress {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetModuleAddress", arg0)
	ret0, _ := ret[0].(types.AccAddress)
	return ret0
}

// GetModuleAddress indicates an expected call of GetModuleAddress.
func (mr *MockFeeCollectorMockRecorder) GetModuleAddress(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetModuleAddress", reflect.TypeOf((*MockFeeCollector)(nil).GetModuleAddress), arg0)
}

// SendCoinsFromAccountToModule mocks base method.
func (m *MockFeeCollector) SendCoinsFromAccountToModule(arg0 types.Context, arg1 types.AccAddress, arg2 string, arg3 types.Coins) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendCoinsFromAccountToModule", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendCoinsFromAccountToModule indicates an expected call of SendCoinsFromAccountToModule.
func (mr *MockFeeCollectorMockRecorder) SendCoinsFromAccountToModule(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendCoinsFromAccountToModule", reflect.TypeOf((*MockFeeCollector)(nil).SendCoinsFromAccountToModule), arg0, arg1, arg2, arg3)
}
