// Code generated by MockGen. DO NOT EDIT.
// Source: consensus/tendermint/validator/validator_interface.go

// Package validator is a generated GoMock package.
package validator

import (
	common "github.com/clearmatics/autonity/common"
	config "github.com/clearmatics/autonity/consensus/tendermint/config"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockValidator is a mock of Validator interface
type MockValidator struct {
	ctrl     *gomock.Controller
	recorder *MockValidatorMockRecorder
}

// MockValidatorMockRecorder is the mock recorder for MockValidator
type MockValidatorMockRecorder struct {
	mock *MockValidator
}

// NewMockValidator creates a new mock instance
func NewMockValidator(ctrl *gomock.Controller) *MockValidator {
	mock := &MockValidator{ctrl: ctrl}
	mock.recorder = &MockValidatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockValidator) EXPECT() *MockValidatorMockRecorder {
	return m.recorder
}

// Address mocks base method
func (m *MockValidator) Address() common.Address {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Address")
	ret0, _ := ret[0].(common.Address)
	return ret0
}

// Address indicates an expected call of Address
func (mr *MockValidatorMockRecorder) Address() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Address", reflect.TypeOf((*MockValidator)(nil).Address))
}

// String mocks base method
func (m *MockValidator) String() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String
func (mr *MockValidatorMockRecorder) String() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*MockValidator)(nil).String))
}

// MockSet is a mock of Set interface
type MockSet struct {
	ctrl     *gomock.Controller
	recorder *MockSetMockRecorder
}

// MockSetMockRecorder is the mock recorder for MockSet
type MockSetMockRecorder struct {
	mock *MockSet
}

// NewMockSet creates a new mock instance
func NewMockSet(ctrl *gomock.Controller) *MockSet {
	mock := &MockSet{ctrl: ctrl}
	mock.recorder = &MockSetMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSet) EXPECT() *MockSetMockRecorder {
	return m.recorder
}

// CalcProposer mocks base method
func (m *MockSet) CalcProposer(lastProposer common.Address, round uint64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CalcProposer", lastProposer, round)
}

// CalcProposer indicates an expected call of CalcProposer
func (mr *MockSetMockRecorder) CalcProposer(lastProposer, round interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CalcProposer", reflect.TypeOf((*MockSet)(nil).CalcProposer), lastProposer, round)
}

// Size mocks base method
func (m *MockSet) Size() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Size")
	ret0, _ := ret[0].(int)
	return ret0
}

// Size indicates an expected call of Size
func (mr *MockSetMockRecorder) Size() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Size", reflect.TypeOf((*MockSet)(nil).Size))
}

// List mocks base method
func (m *MockSet) List() []Validator {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List")
	ret0, _ := ret[0].([]Validator)
	return ret0
}

// List indicates an expected call of List
func (mr *MockSetMockRecorder) List() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockSet)(nil).List))
}

// GetByIndex mocks base method
func (m *MockSet) GetByIndex(i uint64) Validator {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByIndex", i)
	ret0, _ := ret[0].(Validator)
	return ret0
}

// GetByIndex indicates an expected call of GetByIndex
func (mr *MockSetMockRecorder) GetByIndex(i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByIndex", reflect.TypeOf((*MockSet)(nil).GetByIndex), i)
}

// GetByAddress mocks base method
func (m *MockSet) GetByAddress(addr common.Address) (int, Validator) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByAddress", addr)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(Validator)
	return ret0, ret1
}

// GetByAddress indicates an expected call of GetByAddress
func (mr *MockSetMockRecorder) GetByAddress(addr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByAddress", reflect.TypeOf((*MockSet)(nil).GetByAddress), addr)
}

// GetProposer mocks base method
func (m *MockSet) GetProposer() Validator {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProposer")
	ret0, _ := ret[0].(Validator)
	return ret0
}

// GetProposer indicates an expected call of GetProposer
func (mr *MockSetMockRecorder) GetProposer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProposer", reflect.TypeOf((*MockSet)(nil).GetProposer))
}

// IsProposer mocks base method
func (m *MockSet) IsProposer(address common.Address) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsProposer", address)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsProposer indicates an expected call of IsProposer
func (mr *MockSetMockRecorder) IsProposer(address interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsProposer", reflect.TypeOf((*MockSet)(nil).IsProposer), address)
}

// AddValidator mocks base method
func (m *MockSet) AddValidator(address common.Address) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddValidator", address)
	ret0, _ := ret[0].(bool)
	return ret0
}

// AddValidator indicates an expected call of AddValidator
func (mr *MockSetMockRecorder) AddValidator(address interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddValidator", reflect.TypeOf((*MockSet)(nil).AddValidator), address)
}

// RemoveValidator mocks base method
func (m *MockSet) RemoveValidator(address common.Address) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveValidator", address)
	ret0, _ := ret[0].(bool)
	return ret0
}

// RemoveValidator indicates an expected call of RemoveValidator
func (mr *MockSetMockRecorder) RemoveValidator(address interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveValidator", reflect.TypeOf((*MockSet)(nil).RemoveValidator), address)
}

// Copy mocks base method
func (m *MockSet) Copy() Set {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Copy")
	ret0, _ := ret[0].(Set)
	return ret0
}

// Copy indicates an expected call of Copy
func (mr *MockSetMockRecorder) Copy() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Copy", reflect.TypeOf((*MockSet)(nil).Copy))
}

// F mocks base method
func (m *MockSet) F() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "F")
	ret0, _ := ret[0].(int)
	return ret0
}

// F indicates an expected call of F
func (mr *MockSetMockRecorder) F() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "F", reflect.TypeOf((*MockSet)(nil).F))
}

// Quorum mocks base method
func (m *MockSet) Quorum() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Quorum")
	ret0, _ := ret[0].(int)
	return ret0
}

// Quorum indicates an expected call of Quorum
func (mr *MockSetMockRecorder) Quorum() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Quorum", reflect.TypeOf((*MockSet)(nil).Quorum))
}

// Policy mocks base method
func (m *MockSet) Policy() config.ProposerPolicy {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Policy")
	ret0, _ := ret[0].(config.ProposerPolicy)
	return ret0
}

// Policy indicates an expected call of Policy
func (mr *MockSetMockRecorder) Policy() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Policy", reflect.TypeOf((*MockSet)(nil).Policy))
}
