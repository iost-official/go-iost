// Code generated by MockGen. DO NOT EDIT.
// Source: interface.go

// Package txpool_mock is a generated GoMock package.
package txpool_mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	block "github.com/iost-official/go-iost/v3/core/block"
	blockcache "github.com/iost-official/go-iost/v3/core/blockcache"
	tx "github.com/iost-official/go-iost/v3/core/tx"
	txpool "github.com/iost-official/go-iost/v3/core/txpool"
)

// MockTxPool is a mock of TxPool interface.
type MockTxPool struct {
	ctrl     *gomock.Controller
	recorder *MockTxPoolMockRecorder
}

// MockTxPoolMockRecorder is the mock recorder for MockTxPool.
type MockTxPoolMockRecorder struct {
	mock *MockTxPool
}

// NewMockTxPool creates a new mock instance.
func NewMockTxPool(ctrl *gomock.Controller) *MockTxPool {
	mock := &MockTxPool{ctrl: ctrl}
	mock.recorder = &MockTxPoolMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTxPool) EXPECT() *MockTxPoolMockRecorder {
	return m.recorder
}

// AddLinkedNode mocks base method.
func (m *MockTxPool) AddLinkedNode(linkedNode *blockcache.BlockCacheNode) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddLinkedNode", linkedNode)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddLinkedNode indicates an expected call of AddLinkedNode.
func (mr *MockTxPoolMockRecorder) AddLinkedNode(linkedNode interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddLinkedNode", reflect.TypeOf((*MockTxPool)(nil).AddLinkedNode), linkedNode)
}

// AddTx mocks base method.
func (m *MockTxPool) AddTx(tx *tx.Tx, from string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddTx", tx, from)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddTx indicates an expected call of AddTx.
func (mr *MockTxPoolMockRecorder) AddTx(tx, from interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddTx", reflect.TypeOf((*MockTxPool)(nil).AddTx), tx, from)
}

// Close mocks base method.
func (m *MockTxPool) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockTxPoolMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockTxPool)(nil).Close))
}

// DelTx mocks base method.
func (m *MockTxPool) DelTx(hash []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DelTx", hash)
	ret0, _ := ret[0].(error)
	return ret0
}

// DelTx indicates an expected call of DelTx.
func (mr *MockTxPoolMockRecorder) DelTx(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DelTx", reflect.TypeOf((*MockTxPool)(nil).DelTx), hash)
}

// ExistTxs mocks base method.
func (m *MockTxPool) ExistTxs(hash []byte, chainBlock *block.Block) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExistTxs", hash, chainBlock)
	ret0, _ := ret[0].(bool)
	return ret0
}

// ExistTxs indicates an expected call of ExistTxs.
func (mr *MockTxPoolMockRecorder) ExistTxs(hash, chainBlock interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExistTxs", reflect.TypeOf((*MockTxPool)(nil).ExistTxs), hash, chainBlock)
}

// GetFromChain mocks base method.
func (m *MockTxPool) GetFromChain(hash []byte) (*tx.Tx, *tx.TxReceipt, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFromChain", hash)
	ret0, _ := ret[0].(*tx.Tx)
	ret1, _ := ret[1].(*tx.TxReceipt)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetFromChain indicates an expected call of GetFromChain.
func (mr *MockTxPoolMockRecorder) GetFromChain(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFromChain", reflect.TypeOf((*MockTxPool)(nil).GetFromChain), hash)
}

// GetFromPending mocks base method.
func (m *MockTxPool) GetFromPending(hash []byte) (*tx.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFromPending", hash)
	ret0, _ := ret[0].(*tx.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFromPending indicates an expected call of GetFromPending.
func (mr *MockTxPoolMockRecorder) GetFromPending(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFromPending", reflect.TypeOf((*MockTxPool)(nil).GetFromPending), hash)
}

// PendingTx mocks base method.
func (m *MockTxPool) PendingTx() (*txpool.SortedTxMap, *blockcache.BlockCacheNode) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PendingTx")
	ret0, _ := ret[0].(*txpool.SortedTxMap)
	ret1, _ := ret[1].(*blockcache.BlockCacheNode)
	return ret0, ret1
}

// PendingTx indicates an expected call of PendingTx.
func (mr *MockTxPoolMockRecorder) PendingTx() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PendingTx", reflect.TypeOf((*MockTxPool)(nil).PendingTx))
}
