package v8

/*
#include <stdlib.h>
#include "v8/vm.h"
char *requireModule(SandboxPtr, const char *);
int goTransfer(SandboxPtr, const char *, const char *, const char *, size_t *);
int goWithdraw(SandboxPtr, const char *, const char *, size_t *);
int goDeposit(SandboxPtr, const char *, const char *, size_t *);
int goTopUp(SandboxPtr, const char *, const char *, const char *, size_t *);
int goCountermand(SandboxPtr, const char *, const char *, const char *, size_t *);
char *goBlockInfo(SandboxPtr, size_t *);
char *goTxInfo(SandboxPtr, size_t *);
char *goCall(SandboxPtr, const char *, const char *, const char *, size_t *);
int goPut(SandboxPtr, const char *, const char *, size_t *);
char *goGet(SandboxPtr, const char *, size_t *);
int goDel(SandboxPtr, const char *, size_t *);
char *goGlobalGet(SandboxPtr, const char *, const char *, size_t *);
*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"unsafe"

	"strings"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/host"
)

// A Sandbox is an execution environment that allows separate, unrelated, JavaScript
// code to run in a single instance of IVM.
type Sandbox struct {
	id      int
	isolate C.IsolatePtr
	context C.SandboxPtr
	modules Modules
	host    *host.Host
	jsPath  string
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
		(C.delFunc)(unsafe.Pointer(C.goDel)),
		(C.globalGetFunc)(unsafe.Pointer(C.goGlobalGet)))
}

func (sbx *Sandbox) SetGasLimit(limit uint64) {
	C.setSandboxGasLimit(sbx.context, C.size_t(limit))
}

func (sbx *Sandbox) SetHost(host *host.Host) {
	sbx.host = host
	sbx.SetGasLimit(uint64(host.GasLimit()))
}

func (sbx *Sandbox) SetModule(name, code string) {
	if name == "" || code == "" {
		return
	}
	m := NewModule(name, code)
	sbx.modules.Set(m)
}

func (sbx *Sandbox) SetJSPath(path string) {
	sbx.jsPath = path
	cPath := C.CString(path)
	C.setJSPath(sbx.context, cPath)
}

func (sbx *Sandbox) Prepare(contract *contract.Contract, function string, args []interface{}) (string, error) {
	name := contract.ID
	code := contract.Code

	sbx.SetModule(name, code)

	if function == "constructor" {
		return fmt.Sprintf(`
var _native_main = require('%s');
var obj = new _native_main();

// store kv that was constructed by contract.
Object.keys(obj).forEach((key) => {
   let val = obj[key];
   IOSTContractStorage.put(key, val);
});
`, name), nil
	}

	argStr, err := formatFuncArgs(args)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`
var _native_main = require('%s');
var obj = new _native_main();

var objObserver = observer.create(obj)

// Object.keys(obj).forEach((key) => {
//     let val = obj[key];
//
//     Object.defineProperty(obj, key, {
//         configurable: false,
//         enumerable: true,
//         get: function() {
//             return IOSTContractStorage.get(key, val);
//         },
//         set: function() {
// 			val = setVal;
// 			IOSTContractStorage.put(key, val);
//         }
//     })
// });

// run contract with specified function and args.
objObserver.%s(%s)
// objObserver['%s'].apply(obj, %v);
`, name, function, strings.Trim(argStr, "[]")), nil
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

func formatFuncArgs(args []interface{}) (string, error) {
	argStr, err := json.Marshal(args)
	if err != nil {
		return "", err
	}

	return string(argStr), nil
}
