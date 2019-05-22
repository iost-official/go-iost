package v8

/*
#include <stdlib.h>
#include "v8/vm.h"
char* goBlockInfo(SandboxPtr, CStr *, size_t *);
char* goTxInfo(SandboxPtr, CStr *, size_t *);
char* goContextInfo(SandboxPtr, CStr *, size_t *);
char* goCall(SandboxPtr, const CStr, const CStr, const CStr, CStr *, size_t *);
char* goCallWithAuth(SandboxPtr, const CStr, const CStr, const CStr, CStr *, size_t *);
char* goRequireAuth(SandboxPtr, const CStr, const CStr, bool *, size_t *);
char* goReceipt(SandboxPtr, const CStr, size_t *);
char* goEvent(SandboxPtr, const CStr, size_t *);

char* goPut(SandboxPtr, const CStr, const CStr, const CStr, size_t *);
char* goHas(SandboxPtr, const CStr, const CStr, bool *, size_t *);
char* goGet(SandboxPtr, const CStr, const CStr, CStr *, size_t *);
char* goDel(SandboxPtr, const CStr, const CStr, size_t *);
char* goMapPut(SandboxPtr, const CStr, const CStr, const CStr, const CStr, size_t *);
char* goMapHas(SandboxPtr, const CStr, const CStr, const CStr, bool *, size_t *);
char* goMapGet(SandboxPtr, const CStr, const CStr, const CStr, CStr *, size_t *);
char* goMapDel(SandboxPtr, const CStr, const CStr, const CStr, size_t *);
char* goMapKeys(SandboxPtr, const CStr, const CStr, CStr *, size_t *);
char* goMapLen(SandboxPtr, const CStr, const CStr, size_t *, size_t *);

char* goGlobalHas(SandboxPtr, const CStr, const CStr, const CStr, bool *, size_t *);
char* goGlobalGet(SandboxPtr, const CStr, const CStr, const CStr, CStr *, size_t *);
char* goGlobalMapHas(SandboxPtr, const CStr, const CStr, const CStr, const CStr, bool *, size_t *);
char* goGlobalMapGet(SandboxPtr, const CStr, const CStr, const CStr, const CStr, CStr *, size_t *);
char* goGlobalMapKeys(SandboxPtr, const CStr,  const CStr, const CStr, CStr *, size_t *);
char* goGlobalMapLen(SandboxPtr, const CStr, const CStr, const CStr, size_t *, size_t *);

char* goConsoleLog(SandboxPtr, const CStr, const CStr);

CStr goSha3(SandboxPtr, const CStr, size_t *);
int goVerify(SandboxPtr, const CStr, const CStr, const CStr, const CStr, size_t *);
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

const resultMaxLength = 65536 // byte

// Error message
var (
	ErrResultTooLong = errors.New("result too long")
)

// Sandbox is an execution environment that allows separate, unrelated, JavaScript
// code to run in a single instance of IVM.
type Sandbox struct {
	id      int
	isolate C.IsolateWrapperPtr
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
	C.InitGoCrypto((C.sha3Func)(C.goSha3), (C.verifyFunc)(C.goVerify))
	C.loadVM(sbx.context, C.int(vmType))
}

// SetGasLimit set gas limit in context
func (sbx *Sandbox) SetGasLimit(limit int64) {
	C.setSandboxGasLimit(sbx.context, C.size_t(limit))
}

// SetHost set host in sandbox and set gas limit
func (sbx *Sandbox) SetHost(host *host.Host) {
	sbx.host = host
	sbx.SetGasLimit(host.GasLimitValue())
}

// SetJSPath set js path and ReloadVM
func (sbx *Sandbox) SetJSPath(path string, vmType vmPoolType) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	C.setJSPath(sbx.context, cPath)
	C.loadVM(sbx.context, C.int(vmType))
}

// Validate contract before save, return err if invalid
func (sbx *Sandbox) Validate(contract *contract.Contract) error {
	code := moduleReplacer.Replace(contract.Code)
	cCode := newCStr(code)
	defer C.free(unsafe.Pointer(cCode.data))

	abi, _ := json.Marshal(contract.Info.Abi)
	cAbi := newCStr(string(abi))
	defer C.free(unsafe.Pointer(cAbi.data))

	var (
		cResult C.CStr
		cErrMsg C.CStr
	)
	ret := C.validate(sbx.context, cCode, cAbi, &cResult, &cErrMsg)

	result := cResult.GoString()
	C.free(unsafe.Pointer(cResult.data))

	if ret == 1 || result != "success" {
		errMsg := cErrMsg.GoString()
		C.free(unsafe.Pointer(cErrMsg.data))
		return fmt.Errorf("validate code error: %v, result: %v", errMsg, result)
	}

	return nil
}

// Compile contract before execution, return compiled code
func (sbx *Sandbox) Compile(contract *contract.Contract) (string, error) {
	code := moduleReplacer.Replace(contract.Code)
	cCode := newCStr(code)
	defer C.free(unsafe.Pointer(cCode.data))

	var (
		cCompiledCode C.CStr
		cErrMsg       C.CStr
	)
	ret := C.compile(sbx.context, cCode, &cCompiledCode, &cErrMsg)
	if ret == 1 {
		errMsg := cErrMsg.GoString()
		C.free(unsafe.Pointer(cErrMsg.data))
		return "", errors.New(errMsg)
	}

	compiledCode := cCompiledCode.GoString()
	C.free(unsafe.Pointer(cCompiledCode.data))

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
const obj = new module.exports;

// run contract with specified function and args
let rs = obj.%s(%s);
if ((typeof rs === 'function') || (typeof rs === 'object')) {
	_IOSTInstruction_counter.incr(12);
	rs = JSON.stringify(rs);
}
rs;
`, code, function, argStr), nil
}

// Execute prepared code, return results, gasUsed
func (sbx *Sandbox) Execute(preparedCode string) (string, int64, error) {
	cCode := newCStr(preparedCode)
	defer C.free(unsafe.Pointer(cCode.data))
	expireTime := C.longlong(sbx.host.Deadline().UnixNano())

	rs := C.Execute(sbx.context, cCode, expireTime)
	defer C.free(unsafe.Pointer(rs.Value.data))
	defer C.free(unsafe.Pointer(rs.Err.data))

	gasUsed := rs.gasUsed

	if rs.Value.size > resultMaxLength {
		return "", int64(gasUsed), ErrResultTooLong
	}

	var result string
	result = rs.Value.GoString()

	var err error
	if rs.Err.data != nil {
		err = errors.New(rs.Err.GoString())
	}

	return result, int64(gasUsed), err
}

func formatFuncArgs(args []interface{}) (string, error) {
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
