package lua

import "github.com/iost-official/prototype/vm"

// Method lua方法接口的实现
type Method struct {
	name string
	inputCount,
	outputCount int
	privilege vm.Privilege
}

func NewMethod(priv vm.Privilege, name string, inputCount, rtnCount int) Method {
	var m Method
	m.name = name
	m.inputCount = inputCount
	m.outputCount = rtnCount
	m.privilege = priv
	return m
}

func (m *Method) Name() string {
	return m.name
}
func (m *Method) InputCount() int {
	return m.inputCount
}
func (m *Method) OutputCount() int {
	return m.outputCount
}
func (m *Method) Privilege() vm.Privilege {
	return m.privilege
}
