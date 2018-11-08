package v8

/*
#include "v8/vm.h"
*/
import "C"
import (
	"encoding/json"

	"github.com/iost-official/go-iost/vm/host"
	"errors"
)

var (
	ErrGetSandbox = errors.New("get sandbox failed.")
	MessageSuccess = ""
)

//export goBlockInfo
func goBlockInfo(cSbx C.SandboxPtr, info **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	blkInfo, cost := sbx.host.BlockInfo()
	*gasUsed = C.size_t(cost.Data)
	*info = C.CString(string(blkInfo))

	return C.Cstring(MessageSuccess)
}

//export goTxInfo
func goTxInfo(cSbx C.SandboxPtr, info **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	txInfo, cost := sbx.host.TxInfo()
	*gasUsed = C.size_t(cost.Data)
	*info = C.CString(string(txInfo))

	return C.Cstring(MessageSuccess)
}

//export goContextInfo
func goContextInfo(cSbx C.SandboxPtr, info **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	ctxInfo, cost := sbx.host.ContextInfo()
	*gasUsed = C.size_t(cost.Data)
	*info = C.CString(string(ctxInfo))

	return C.Cstring(MessageSuccess)
}

//export goCall
func goCall(cSbx C.SandboxPtr, contract, api, args *C.char, result **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	contractStr := C.GoString(contract)
	apiStr := C.GoString(api)
	argsStr := C.GoString(args)

	callRs, cost, err := sbx.host.Call(contractStr, apiStr, argsStr)
	*gasUsed = C.size_t(cost.Data)
	if err != nil {
		return C.Cstring(err.Error())
	}

	rsStr, err := json.Marshal(callRs)
	if err != nil {
		return C.Cstring(host.ErrInvalidData.Error())
	}

	*result = C.CString(string(rsStr))

	return C.Cstring(MessageSuccess)
}

//export goCallWithAuth
func goCallWithAuth(cSbx C.SandboxPtr, contract, api, args *C.char, result **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	contractStr := C.GoString(contract)
	apiStr := C.GoString(api)
	argsStr := C.GoString(args)

	callRs, cost, err := sbx.host.CallWithAuth(contractStr, apiStr, argsStr)
	*gasUsed = C.size_t(cost.Data)
	if err != nil {
		return C.Cstring(err.Error())
	}

	rsStr, err := json.Marshal(callRs)
	if err != nil {
		return C.Cstring(host.ErrInvalidData.Error())
	}

	*result = C.CString(string(rsStr))

	return C.Cstring(MessageSuccess)
}

//export goRequireAuth
func goRequireAuth(cSbx C.SandboxPtr, ID *C.char, permission *C.char, ok *C.bool, gasUsed *C.size_t) *C.char {
	sbx, sbOk := GetSandbox(cSbx)
	if !sbOk {
		return C.Cstring(ErrGetSandbox.Error())
	}

	pubKeyStr := C.GoString(ID)
	permissionStr := C.GoString(permission)

	callOk, RequireAuthCost := sbx.host.RequireAuth(pubKeyStr, permissionStr)

	*ok = C.bool(callOk)

	*gasUsed = C.size_t(RequireAuthCost.Data)

	return C.Cstring(MessageSuccess)
}

//export goReceipt
func goReceipt(cSbx C.SandboxPtr, content *C.char, gasUsed *C.size_t) *C.char {
	sbx, sbOk := GetSandbox(cSbx)
	if !sbOk {
		return C.Cstring(ErrGetSandbox.Error())
	}

	contentStr := C.GoString(content)

	cost := sbx.host.Receipt(contentStr)

	*gasUsed = C.size_t(cost.Data)

	return C.Cstring(MessageSuccess)
}

//export goEvent
func goEvent(cSbx C.SandboxPtr, content *C.char, gasUsed *C.size_t) *C.char {
	sbx, sbOk := GetSandbox(cSbx)
	if !sbOk {
		return C.Cstring(ErrGetSandbox.Error())
	}

	contentStr := C.GoString(content)

	cost := sbx.host.PostEvent(contentStr)

	*gasUsed = C.size_t(cost.CPU)

	return C.Cstring(MessageSuccess)
}
