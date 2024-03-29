// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/iost-official/go-iost/v3/p2p (interfaces: Service)

// Package p2p_mock is a generated GoMock package.
package p2p_mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	p2p "github.com/iost-official/go-iost/v3/p2p"
	peer "github.com/libp2p/go-libp2p/core/peer"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// Broadcast mocks base method.
func (m *MockService) Broadcast(arg0 []byte, arg1 p2p.MessageType, arg2 p2p.MessagePriority) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Broadcast", arg0, arg1, arg2)
}

// Broadcast indicates an expected call of Broadcast.
func (mr *MockServiceMockRecorder) Broadcast(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Broadcast", reflect.TypeOf((*MockService)(nil).Broadcast), arg0, arg1, arg2)
}

// ConnectBPs mocks base method.
func (m *MockService) ConnectBPs(arg0 []string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ConnectBPs", arg0)
}

// ConnectBPs indicates an expected call of ConnectBPs.
func (mr *MockServiceMockRecorder) ConnectBPs(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConnectBPs", reflect.TypeOf((*MockService)(nil).ConnectBPs), arg0)
}

// Deregister mocks base method.
func (m *MockService) Deregister(arg0 string, arg1 ...p2p.MessageType) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Deregister", varargs...)
}

// Deregister indicates an expected call of Deregister.
func (mr *MockServiceMockRecorder) Deregister(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Deregister", reflect.TypeOf((*MockService)(nil).Deregister), varargs...)
}

// GetAllNeighbors mocks base method.
func (m *MockService) GetAllNeighbors() []*p2p.Peer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllNeighbors")
	ret0, _ := ret[0].([]*p2p.Peer)
	return ret0
}

// GetAllNeighbors indicates an expected call of GetAllNeighbors.
func (mr *MockServiceMockRecorder) GetAllNeighbors() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllNeighbors", reflect.TypeOf((*MockService)(nil).GetAllNeighbors))
}

// ID mocks base method.
func (m *MockService) ID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ID")
	ret0, _ := ret[0].(string)
	return ret0
}

// ID indicates an expected call of ID.
func (mr *MockServiceMockRecorder) ID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ID", reflect.TypeOf((*MockService)(nil).ID))
}

// PutPeerToBlack mocks base method.
func (m *MockService) PutPeerToBlack(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PutPeerToBlack", arg0)
}

// PutPeerToBlack indicates an expected call of PutPeerToBlack.
func (mr *MockServiceMockRecorder) PutPeerToBlack(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutPeerToBlack", reflect.TypeOf((*MockService)(nil).PutPeerToBlack), arg0)
}

// Register mocks base method.
func (m *MockService) Register(arg0 string, arg1 ...p2p.MessageType) chan p2p.IncomingMessage {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Register", varargs...)
	ret0, _ := ret[0].(chan p2p.IncomingMessage)
	return ret0
}

// Register indicates an expected call of Register.
func (mr *MockServiceMockRecorder) Register(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockService)(nil).Register), varargs...)
}

// SendToPeer mocks base method.
func (m *MockService) SendToPeer(arg0 peer.ID, arg1 []byte, arg2 p2p.MessageType, arg3 p2p.MessagePriority) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendToPeer", arg0, arg1, arg2, arg3)
}

// SendToPeer indicates an expected call of SendToPeer.
func (mr *MockServiceMockRecorder) SendToPeer(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendToPeer", reflect.TypeOf((*MockService)(nil).SendToPeer), arg0, arg1, arg2, arg3)
}

// Start mocks base method.
func (m *MockService) Start() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockServiceMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockService)(nil).Start))
}

// Stop mocks base method.
func (m *MockService) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop.
func (mr *MockServiceMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockService)(nil).Stop))
}
