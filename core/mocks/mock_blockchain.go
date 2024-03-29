// Code generated by MockGen. DO NOT EDIT.
// Source: interface.go

// Package core_mock is a generated GoMock package.
package core_mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	block "github.com/iost-official/go-iost/v3/core/block"
	tx "github.com/iost-official/go-iost/v3/core/tx"
)

// MockChain is a mock of Chain interface.
type MockChain struct {
	ctrl     *gomock.Controller
	recorder *MockChainMockRecorder
}

// MockChainMockRecorder is the mock recorder for MockChain.
type MockChainMockRecorder struct {
	mock *MockChain
}

// NewMockChain creates a new mock instance.
func NewMockChain(ctrl *gomock.Controller) *MockChain {
	mock := &MockChain{ctrl: ctrl}
	mock.recorder = &MockChainMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockChain) EXPECT() *MockChainMockRecorder {
	return m.recorder
}

// CheckLength mocks base method.
func (m *MockChain) CheckLength() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CheckLength")
}

// CheckLength indicates an expected call of CheckLength.
func (mr *MockChainMockRecorder) CheckLength() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckLength", reflect.TypeOf((*MockChain)(nil).CheckLength))
}

// Close mocks base method.
func (m *MockChain) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockChainMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockChain)(nil).Close))
}

// Draw mocks base method.
func (m *MockChain) Draw(arg0, arg1 int64) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Draw", arg0, arg1)
	ret0, _ := ret[0].(string)
	return ret0
}

// Draw indicates an expected call of Draw.
func (mr *MockChainMockRecorder) Draw(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Draw", reflect.TypeOf((*MockChain)(nil).Draw), arg0, arg1)
}

// GetBlockByHash mocks base method.
func (m *MockChain) GetBlockByHash(blockHash []byte) (*block.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockByHash", blockHash)
	ret0, _ := ret[0].(*block.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockByHash indicates an expected call of GetBlockByHash.
func (mr *MockChainMockRecorder) GetBlockByHash(blockHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockByHash", reflect.TypeOf((*MockChain)(nil).GetBlockByHash), blockHash)
}

// GetBlockByNumber mocks base method.
func (m *MockChain) GetBlockByNumber(number int64) (*block.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockByNumber", number)
	ret0, _ := ret[0].(*block.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockByNumber indicates an expected call of GetBlockByNumber.
func (mr *MockChainMockRecorder) GetBlockByNumber(number interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockByNumber", reflect.TypeOf((*MockChain)(nil).GetBlockByNumber), number)
}

// GetBlockNumberByTxHash mocks base method.
func (m *MockChain) GetBlockNumberByTxHash(hash []byte) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockNumberByTxHash", hash)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockNumberByTxHash indicates an expected call of GetBlockNumberByTxHash.
func (mr *MockChainMockRecorder) GetBlockNumberByTxHash(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockNumberByTxHash", reflect.TypeOf((*MockChain)(nil).GetBlockNumberByTxHash), hash)
}

// GetHashByNumber mocks base method.
func (m *MockChain) GetHashByNumber(number int64) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHashByNumber", number)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHashByNumber indicates an expected call of GetHashByNumber.
func (mr *MockChainMockRecorder) GetHashByNumber(number interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHashByNumber", reflect.TypeOf((*MockChain)(nil).GetHashByNumber), number)
}

// GetReceipt mocks base method.
func (m *MockChain) GetReceipt(Hash []byte) (*tx.TxReceipt, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetReceipt", Hash)
	ret0, _ := ret[0].(*tx.TxReceipt)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetReceipt indicates an expected call of GetReceipt.
func (mr *MockChainMockRecorder) GetReceipt(Hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetReceipt", reflect.TypeOf((*MockChain)(nil).GetReceipt), Hash)
}

// GetReceiptByTxHash mocks base method.
func (m *MockChain) GetReceiptByTxHash(Hash []byte) (*tx.TxReceipt, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetReceiptByTxHash", Hash)
	ret0, _ := ret[0].(*tx.TxReceipt)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetReceiptByTxHash indicates an expected call of GetReceiptByTxHash.
func (mr *MockChainMockRecorder) GetReceiptByTxHash(Hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetReceiptByTxHash", reflect.TypeOf((*MockChain)(nil).GetReceiptByTxHash), Hash)
}

// GetTx mocks base method.
func (m *MockChain) GetTx(hash []byte) (*tx.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTx", hash)
	ret0, _ := ret[0].(*tx.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTx indicates an expected call of GetTx.
func (mr *MockChainMockRecorder) GetTx(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTx", reflect.TypeOf((*MockChain)(nil).GetTx), hash)
}

// HasReceipt mocks base method.
func (m *MockChain) HasReceipt(hash []byte) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasReceipt", hash)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasReceipt indicates an expected call of HasReceipt.
func (mr *MockChainMockRecorder) HasReceipt(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasReceipt", reflect.TypeOf((*MockChain)(nil).HasReceipt), hash)
}

// HasTx mocks base method.
func (m *MockChain) HasTx(hash []byte) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasTx", hash)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasTx indicates an expected call of HasTx.
func (mr *MockChainMockRecorder) HasTx(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasTx", reflect.TypeOf((*MockChain)(nil).HasTx), hash)
}

// Length mocks base method.
func (m *MockChain) Length() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Length")
	ret0, _ := ret[0].(int64)
	return ret0
}

// Length indicates an expected call of Length.
func (mr *MockChainMockRecorder) Length() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Length", reflect.TypeOf((*MockChain)(nil).Length))
}

// Push mocks base method.
func (m *MockChain) Push(block *block.Block) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Push", block)
	ret0, _ := ret[0].(error)
	return ret0
}

// Push indicates an expected call of Push.
func (mr *MockChainMockRecorder) Push(block interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Push", reflect.TypeOf((*MockChain)(nil).Push), block)
}

// SetLength mocks base method.
func (m *MockChain) SetLength(i int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetLength", i)
}

// SetLength indicates an expected call of SetLength.
func (mr *MockChainMockRecorder) SetLength(i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetLength", reflect.TypeOf((*MockChain)(nil).SetLength), i)
}

// Size mocks base method.
func (m *MockChain) Size() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Size")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Size indicates an expected call of Size.
func (mr *MockChainMockRecorder) Size() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Size", reflect.TypeOf((*MockChain)(nil).Size))
}

// Top mocks base method.
func (m *MockChain) Top() (*block.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Top")
	ret0, _ := ret[0].(*block.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Top indicates an expected call of Top.
func (mr *MockChainMockRecorder) Top() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Top", reflect.TypeOf((*MockChain)(nil).Top))
}

// TxTotal mocks base method.
func (m *MockChain) TxTotal() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TxTotal")
	ret0, _ := ret[0].(int64)
	return ret0
}

// TxTotal indicates an expected call of TxTotal.
func (mr *MockChainMockRecorder) TxTotal() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TxTotal", reflect.TypeOf((*MockChain)(nil).TxTotal))
}
