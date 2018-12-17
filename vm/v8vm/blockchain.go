package v8

/*
#include "v8/vm.h"
*/
import "C"
import (
	"encoding/json"
	"errors"

	"github.com/iost-official/go-iost/vm/host"
)

var (
	// ErrGetSandbox error when GetSandbox failed.
	ErrGetSandbox = errors.New("get sandbox failed")
)

func newCStr(str string) C.CStr {
	cstr := C.CStr{}
	cstr.data = C.CString(str)
	cstr.size = C.int(len(str))
	return cstr
}

func (cstr *C.CStr) SetString(str string) {
	cstr.data = C.CString(str)
	cstr.size = C.int(len(str))
	return
}

func (cstr C.CStr) GoString() string {
	return C.GoStringN(cstr.data, cstr.size)
}

//export goBlockInfo
func goBlockInfo(cSbx C.SandboxPtr, info *C.CStr, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	blkInfo, cost := sbx.host.BlockInfo()
	*gasUsed = C.size_t(cost.CPU)
	info.SetString(string(blkInfo))

	return nil
}

//export goTxInfo
func goTxInfo(cSbx C.SandboxPtr, info *C.CStr, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	txInfo, cost := sbx.host.TxInfo()
	*gasUsed = C.size_t(cost.CPU)
	info.SetString(string(txInfo))

	return nil
}

//export goContextInfo
func goContextInfo(cSbx C.SandboxPtr, info *C.CStr, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	ctxInfo, cost := sbx.host.ContextInfo()
	*gasUsed = C.size_t(cost.CPU)
	info.SetString(string(ctxInfo))

	return nil
}

//export goCall
func goCall(cSbx C.SandboxPtr, contract, api, args C.CStr, result *C.CStr, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	contractStr := contract.GoString()
	apiStr := api.GoString()
	argsStr := args.GoString()

	callRs, cost, err := sbx.host.Call(contractStr, apiStr, argsStr)
	*gasUsed = C.size_t(cost.CPU)
	if err != nil {
		return C.CString(err.Error())
	}

	rsStr, err := json.Marshal(callRs)
	if err != nil {
		return C.CString(host.ErrInvalidData.Error())
	}

	result.SetString(string(rsStr))

	return nil
}

//export goCallWithAuth
func goCallWithAuth(cSbx C.SandboxPtr, contract, api, args C.CStr, result *C.CStr, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	contractStr := contract.GoString()
	apiStr := api.GoString()
	argsStr := args.GoString()

	callRs, cost, err := sbx.host.CallWithAuth(contractStr, apiStr, argsStr)
	*gasUsed = C.size_t(cost.CPU)
	if err != nil {
		return C.CString(err.Error())
	}

	rsStr, err := json.Marshal(callRs)
	if err != nil {
		return C.CString(host.ErrInvalidData.Error())
	}

	result.SetString(string(rsStr))

	return nil
}

//export goRequireAuth
func goRequireAuth(cSbx C.SandboxPtr, ID, permission C.CStr, ok *C.bool, gasUsed *C.size_t) *C.char {
	sbx, sbOk := GetSandbox(cSbx)
	if !sbOk {
		return C.CString(ErrGetSandbox.Error())
	}

	pubKeyStr := ID.GoString()
	permissionStr := permission.GoString()

	callOk, RequireAuthCost := sbx.host.RequireAuth(pubKeyStr, permissionStr)

	*ok = C.bool(callOk)

	*gasUsed = C.size_t(RequireAuthCost.CPU)

	return nil
}

//export goReceipt
func goReceipt(cSbx C.SandboxPtr, content C.CStr, gasUsed *C.size_t) *C.char {
	sbx, sbOk := GetSandbox(cSbx)
	if !sbOk {
		return C.CString(ErrGetSandbox.Error())
	}

	contentStr := content.GoString()

	cost := sbx.host.Receipt(contentStr)

	*gasUsed = C.size_t(cost.CPU)

	return nil
}

//export goEvent
func goEvent(cSbx C.SandboxPtr, content C.CStr, gasUsed *C.size_t) *C.char {
	sbx, sbOk := GetSandbox(cSbx)
	if !sbOk {
		return C.CString(ErrGetSandbox.Error())
	}

	contentStr := content.GoString()

	cost := sbx.host.PostEvent(contentStr)

	*gasUsed = C.size_t(cost.CPU)

	return nil
}
