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
int goBlockInfo(SandboxPtr, char **, size_t *);
int goTxInfo(SandboxPtr, char **, size_t *);
int goCall(SandboxPtr, const char *, const char *, const char *, char **, size_t *);
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
	"github.com/iost-official/Go-IOS-Protocol/vm/host"
)

// Sandbox is an execution environment that allows separate, unrelated, JavaScript
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

// GetSandbox from sandbox map by sandbox ptr
func GetSandbox(cSbx C.SandboxPtr) (*Sandbox, bool) {
	sbx, ok := sbxMap[cSbx]
	return sbx, ok
}

// NewSandbox generate new sandbox for VM and insert into sandbox map
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

// Release release sandbox and delete from map
func (sbx *Sandbox) Release() {
	if sbx.context != nil {
		delete(sbxMap, sbx.context)
		C.releaseSandbox(sbx.context)
	}
	sbx.context = nil
}

// nolint
// Init add system functions
func (sbx *Sandbox) Init() {
	// init require
	C.InitGoRequire((C.requireFunc)(C.requireModule))
	C.InitGoBlockchain((C.transferFunc)(C.goTransfer),
		(C.withdrawFunc)(C.goWithdraw),
		(C.depositFunc)(C.goDeposit),
		(C.topUpFunc)(C.goTopUp),
		(C.countermandFunc)(C.goCountermand),
		(C.blockInfoFunc)(C.goBlockInfo),
		(C.txInfoFunc)(C.goTxInfo),
		(C.callFunc)(C.goCall))
	C.InitGoStorage((C.putFunc)(C.goPut),
		(C.getFunc)(C.goGet),
		(C.delFunc)(C.goDel),
		(C.globalGetFunc)(C.goGlobalGet))
}

// SetGasLimit set gas limit in context
func (sbx *Sandbox) SetGasLimit(limit int64) {
	C.setSandboxGasLimit(sbx.context, C.size_t(limit))
}

// SetHost set host in sandbox and set gas limit
func (sbx *Sandbox) SetHost(host *host.Host) {
	sbx.host = host
	sbx.SetGasLimit(host.GasLimit())
}

// SetModule generate new module and set module in sandbox
func (sbx *Sandbox) SetModule(name, code string) {
	if name == "" || code == "" {
		return
	}
	m := NewModule(name, code)
	sbx.modules.Set(m)
}

// SetJSPath set js path
func (sbx *Sandbox) SetJSPath(path string) {
	sbx.jsPath = path
	cPath := C.CString(path)
	C.setJSPath(sbx.context, cPath)
}

// Prepare for contract, inject code
func (sbx *Sandbox) Prepare(contract *contract.Contract, function string, args []interface{}) (string, error) {
	name := contract.ID
	code := contract.Code

	sbx.SetModule(name, code)

	if function == "constructor" {
		return fmt.Sprintf(`
var _native_main = require('%s');
var obj = new _native_main();

var ret = 0;
// store kv that was constructed by contract.
Object.keys(obj).forEach((key) => {
   let val = obj[key];
   ret = IOSTContractStorage.put(key, val);
});
ret;
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

// run contract with specified function and args
objObserver.%s(%s)
`, name, function, strings.Trim(argStr, "[]")), nil
}

// Execute prepared code, return results, gasUsed
func (sbx *Sandbox) Execute(preparedCode string) (string, int64, error) {
	cCode := C.CString(preparedCode)
	defer C.free(unsafe.Pointer(cCode))

	rs := C.Execute(sbx.context, cCode)

	result := C.GoString(rs.Value)
	defer C.free(unsafe.Pointer(rs.Value))
	defer C.free(unsafe.Pointer(rs.Err))

	var err error
	if rs.Err != nil {
		err = errors.New(C.GoString(rs.Err))
	}

	gasUsed := rs.gasUsed

	return result, int64(gasUsed), err
}

func formatFuncArgs(args []interface{}) (string, error) {
	argStr, err := json.Marshal(args)
	if err != nil {
		return "", err
	}

	return string(argStr), nil
}
