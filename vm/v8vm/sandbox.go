package v8

/*
#include <stdlib.h>
#include "v8/vm.h"
char* goBlockInfo(SandboxPtr, char **, size_t *);
char* goTxInfo(SandboxPtr, char **, size_t *);
char* goContextInfo(SandboxPtr, char **, size_t *);
char* goCall(SandboxPtr, const char *, const char *, const char *, char **, size_t *);
char* goCallWithAuth(SandboxPtr, const char *, const char *, const char *, char **, size_t *);
char* goRequireAuth(SandboxPtr, const char *, const char *, bool *, size_t *);
char* goReceipt(SandboxPtr, const char *, size_t *);
char* goEvent(SandboxPtr, const char *, size_t *);

char* goPut(SandboxPtr, const char *, const char *, const char *, size_t *);
char* goHas(SandboxPtr, const char *, const char *, bool *, size_t *);
char* goGet(SandboxPtr, const char *, const char *, char **, size_t *);
char* goDel(SandboxPtr, const char *, const char *, size_t *);
char* goMapPut(SandboxPtr, const char *, const char *, const char *, const char *, size_t *);
char* goMapHas(SandboxPtr, const char *, const char *, const char *, bool *, size_t *);
char* goMapGet(SandboxPtr, const char *, const char *, const char *, char **, size_t *);
char* goMapDel(SandboxPtr, const char *, const char *, const char *, size_t *);
char* goMapKeys(SandboxPtr, const char *, const char *, char **, size_t *);
char* goMapLen(SandboxPtr, const char *, const char *, size_t *, size_t *);

char* goGlobalHas(SandboxPtr, const char *, const char *, const char *, bool *, size_t *);
char* goGlobalGet(SandboxPtr, const char *, const char *, const char *, char **, size_t *);
char* goGlobalMapHas(SandboxPtr, const char *, const char *, const char *, const char *, bool *, size_t *);
char* goGlobalMapGet(SandboxPtr, const char *, const char *, const char *, const char *, char **, size_t *);
char* goGlobalMapKeys(SandboxPtr, const char *,  const char *, const char *, char **, size_t *);
char* goGlobalMapLen(SandboxPtr, const char *, const char *, const char *, size_t *, size_t *);

char* goConsoleLog(SandboxPtr, const char *, const char *);
*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"unsafe"

	"strings"

	"sync"

	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/host"
)

// Sandbox is an execution environment that allows separate, unrelated, JavaScript
// code to run in a single instance of IVM.
type Sandbox struct {
	id      int
	isolate C.IsolatePtr
	context C.SandboxPtr
	host    *host.Host
}

//var sbxMap = make(map[C.SandboxPtr]*Sandbox)
var sbxMap sync.Map

// GetSandbox from sandbox map by sandbox ptr
func GetSandbox(cSbx C.SandboxPtr) (*Sandbox, bool) {
	valInterface, ok := sbxMap.Load(cSbx)
	if !ok {
		return nil, ok
	}
	sbx, ok := valInterface.(*Sandbox)
	return sbx, ok
}

// NewSandbox generate new sandbox for VM and insert into sandbox map
func NewSandbox(e *VM) *Sandbox {
	cPath := C.CString(e.jsPath)
	defer C.free(unsafe.Pointer(cPath))
	cSbx := C.newSandbox(e.isolate)
	C.setJSPath(cSbx, cPath)

	s := &Sandbox{
		isolate: e.isolate,
		context: cSbx,
	}
	s.Init(e.vmType)
	sbxMap.Store(cSbx, s)

	return s
}

// Release release sandbox and delete from map
func (sbx *Sandbox) Release() {
	if sbx.context != nil {
		sbxMap.Delete(sbx.context)
		C.releaseSandbox(sbx.context)
	}
	sbx.context = nil
}

// Init add system functions
func (sbx *Sandbox) Init(vmType vmPoolType) {
	// init require
	C.InitGoConsole((C.consoleFunc)(C.goConsoleLog))
	C.InitGoBlockchain(
		(C.blockInfoFunc)(C.goBlockInfo),
		(C.txInfoFunc)(C.goTxInfo),
		(C.contextInfoFunc)(C.goContextInfo),
		(C.callFunc)(C.goCall),
		(C.callFunc)(C.goCallWithAuth),
		(C.requireAuthFunc)(C.goRequireAuth),
		(C.receiptFunc)(C.goReceipt),
		(C.eventFunc)(C.goEvent),
	)
	C.InitGoStorage(
		(C.putFunc)(C.goPut),
		(C.hasFunc)(C.goHas),
		(C.getFunc)(C.goGet),
		(C.delFunc)(C.goDel),
		(C.mapPutFunc)(C.goMapPut),
		(C.mapHasFunc)(C.goMapHas),
		(C.mapGetFunc)(C.goMapGet),
		(C.mapDelFunc)(C.goMapDel),
		(C.mapKeysFunc)(C.goMapKeys),
		(C.mapLenFunc)(C.goMapLen),

		(C.globalHasFunc)(C.goGlobalHas),
		(C.globalGetFunc)(C.goGlobalGet),
		(C.globalMapHasFunc)(C.goGlobalMapHas),
		(C.globalMapGetFunc)(C.goGlobalMapGet),
		(C.globalMapKeysFunc)(C.goGlobalMapKeys),
		(C.globalMapLenFunc)(C.goGlobalMapLen),
	)
	C.loadVM(sbx.context, C.int(vmType))
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

// SetJSPath set js path and ReloadVM
func (sbx *Sandbox) SetJSPath(path string, vmType vmPoolType) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	C.setJSPath(sbx.context, cPath)
	C.loadVM(sbx.context, C.int(vmType))
}

// Compile contract before execution, return compiled code
func (sbx *Sandbox) Compile(contract *contract.Contract) (string, error) {
	code := moduleReplacer.Replace(contract.Code)
	cCode := C.CString(code)
	defer C.free(unsafe.Pointer(cCode))

	var cCompiledCode *C.char
	ret := C.compile(sbx.context, cCode, &cCompiledCode)
	if ret == 1 {
		return "", errors.New("compile code error")
	}

	compiledCode := C.GoString(cCompiledCode)
	C.free(unsafe.Pointer(cCompiledCode))

	return compiledCode, nil
}

// Prepare for contract, inject code
func (sbx *Sandbox) Prepare(contract *contract.Contract, function string, args []interface{}) (string, error) {
	code := contract.Code

	if function == "constructor" {
		return fmt.Sprintf(`
%s
var obj = new module.exports;

var ret = 0;
// store kv that was constructed by contract.
//Object.keys(obj).forEach((key) => {
//   let val = obj[key];
//   ret = IOSTContractStorage.put(key, val);
//});
ret;
`, code), nil
	}

	argStr, err := formatFuncArgs(args)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`
%s;
var obj = new module.exports;

// run contract with specified function and args
obj.%s(%s)
`, code, function, argStr), nil
}

// Execute prepared code, return results, gasUsed
func (sbx *Sandbox) Execute(preparedCode string) (string, int64, error) {
	cCode := C.CString(preparedCode)
	defer C.free(unsafe.Pointer(cCode))
	expireTime := C.longlong(sbx.host.Deadline().UnixNano())

	rs := C.Execute(sbx.context, cCode, expireTime)

	var result string
	result = C.GoString(rs.Value)
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
	if len(args) == 0 {
		// hack for vm_test param2
		return "null", nil
	}
	var strArgs []string
	for _, arg := range args {
		switch v := arg.(type) {
		case []byte:
			strArgs = append(strArgs, string(v))
		default:
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			strArgs = append(strArgs, string(b))
		}
	}
	argStr := strings.Join(strArgs, ",")

	return argStr, nil
}
