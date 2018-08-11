package v8

/*
#include <stdlib.h>
#include "v8/vm.h"
char *requireModule(SandboxPtr, char *);
int goTransfer(SandboxPtr, char *, char *, char *, size_t *);
int goWithdraw(SandboxPtr, char *, char *, size_t *);
int goDeposit(SandboxPtr, char *, char *, size_t *);
int goTopUp(SandboxPtr, char *, char *, char *, size_t *);
int goCountermand(SandboxPtr, char *, char *, char *, size_t *);
char *goBlockInfo(SandboxPtr, size_t *);
char *goTxInfo(SandboxPtr, size_t *);
char *goCall(SandboxPtr, char *, char *, char *, size_t *);
int goPut(SandboxPtr, char *, char *, size_t *);
char *goGet(SandboxPtr, char *, size_t *);
int goDel(SandboxPtr, char *, size_t *);
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm"
	"encoding/json"
)

// A Sandbox is an execution environment that allows separate, unrelated, JavaScript
// code to run in a single instance of IVM.
type Sandbox struct {
	id      int
	isolate C.IsolatePtr
	context C.SandboxPtr
	modules Modules
	host    *new_vm.Host
}

var sbxMap = make(map[C.SandboxPtr]*Sandbox)

func GetSandbox(cSbx C.SandboxPtr) (*Sandbox, bool) {
	sbx, ok := sbxMap[cSbx]
	return sbx, ok
}

func NewSandbox(e *VM) *Sandbox {
	cSbx := C.newSandbox(e.isolate)
	s := &Sandbox{
		isolate: e.isolate,
		context: cSbx,
		modules: NewModules(),
	}
	s.Init()
	sbxMap[cSbx] = s

	return s
}

func (sbx *Sandbox) Release() {
	if sbx.context != nil {
		delete(sbxMap, sbx.context)
		C.releaseSandbox(sbx.context)
	}
	sbx.context = nil
}

func (sbx *Sandbox) Init() {
	// init require
	C.InitGoRequire((C.requireFunc)(unsafe.Pointer(C.requireModule)))
	C.InitGoBlockchain((C.transferFunc)(unsafe.Pointer(C.goTransfer)),
		(C.withdrawFunc)(unsafe.Pointer(C.goWithdraw)),
		(C.depositFunc)(unsafe.Pointer(C.goDeposit)),
		(C.topUpFunc)(unsafe.Pointer(C.goTopUp)),
		(C.countermandFunc)(unsafe.Pointer(C.goCountermand)),
		(C.blockInfoFunc)(unsafe.Pointer(C.goBlockInfo)),
		(C.txInfoFunc)(unsafe.Pointer(C.goTxInfo)),
		(C.callFunc)(unsafe.Pointer(C.goCall)))
	C.InitGoStorage((C.putFunc)(unsafe.Pointer(C.goPut)),
		(C.getFunc)(unsafe.Pointer(C.goGet)),
		(C.delFunc)(unsafe.Pointer(C.goDel)))
}

func (sbx *Sandbox) SetHost(host *new_vm.Host) {
	sbx.host = host
}

func (sbx *Sandbox) SetModule(name, code string) {
	if name == "" || code == "" {
		return
	}
	m := NewModule(name, code)
	sbx.modules.Set(m)
}

func (sbx *Sandbox) Prepare(contract *contract.Contract, function string, args []interface{}) (string, error) {
	name := contract.ID
	code := contract.Code

	sbx.SetModule(name, code)

	argStr, err := json.Marshal(args)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`
var _native_main = NativeModule.require('%s');
var obj = new _native_main();
obj['%s'].apply(obj, %v);
`, name, function, string(argStr)), nil
}

func (sbx *Sandbox) Execute(preparedCode string) (string, error) {
	cCode := C.CString(preparedCode)
	defer C.free(unsafe.Pointer(cCode))

	rs := C.Execute(sbx.context, cCode)

	result := C.GoString(rs.Value)

	var err error
	if rs.Err != nil {
		err = errors.New(C.GoString(rs.Err))
	}

	if rs.Value != nil {
		C.free(unsafe.Pointer(rs.Value))
	}
	if rs.Err != nil {
		C.free(unsafe.Pointer(rs.Err))
	}

	return result, err
}
