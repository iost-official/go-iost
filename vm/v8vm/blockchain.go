package v8

/*
#include "v8/vm.h"
*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iost-official/go-iost/vm/host"
)

var (
	// ErrGetSandbox error when GetSandbox failed.
	ErrGetSandbox = errors.New("get sandbox failed")
)

//export goBlockInfo
func goBlockInfo(cSbx C.SandboxPtr, info **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	blkInfo, cost := sbx.host.BlockInfo()
	*gasUsed = C.size_t(cost.CPU)
	*info = C.CString(string(blkInfo))

	return nil
}

//export goTxInfo
func goTxInfo(cSbx C.SandboxPtr, info **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	txInfo, cost := sbx.host.TxInfo()
	*gasUsed = C.size_t(cost.CPU)
	*info = C.CString(string(txInfo))

	return nil
}

//export goContextInfo
func goContextInfo(cSbx C.SandboxPtr, info **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	ctxInfo, cost := sbx.host.ContextInfo()
	*gasUsed = C.size_t(cost.CPU)
	*info = C.CString(string(ctxInfo))

	return nil
}

//export goCall
func goCall(cSbx C.SandboxPtr, contract, api, args *C.char, result **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	contractStr := C.GoString(contract)
	apiStr := C.GoString(api)
	argsStr := C.GoString(args)

	callRs, cost, err := sbx.host.Call(contractStr, apiStr, argsStr)
	*gasUsed = C.size_t(cost.CPU)
	if err != nil {
		fmt.Printf("goCall err %v %v %v %v\n", contractStr, apiStr, argsStr, err.Error())
		return C.CString(err.Error())
	}

	rsStr, err := json.Marshal(callRs)
	if err != nil {
		return C.CString(host.ErrInvalidData.Error())
	}

	*result = C.CString(string(rsStr))

	return nil
}

//export goCallWithAuth
func goCallWithAuth(cSbx C.SandboxPtr, contract, api, args *C.char, result **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	contractStr := C.GoString(contract)
	apiStr := C.GoString(api)
	argsStr := C.GoString(args)

	callRs, cost, err := sbx.host.CallWithAuth(contractStr, apiStr, argsStr)
	*gasUsed = C.size_t(cost.CPU)
	if err != nil {
		return C.CString(err.Error())
	}

	rsStr, err := json.Marshal(callRs)
	if err != nil {
		return C.CString(host.ErrInvalidData.Error())
	}

	*result = C.CString(string(rsStr))

	return nil
}

//export goRequireAuth
func goRequireAuth(cSbx C.SandboxPtr, ID *C.char, permission *C.char, ok *C.bool, gasUsed *C.size_t) *C.char {
	sbx, sbOk := GetSandbox(cSbx)
	if !sbOk {
		return C.CString(ErrGetSandbox.Error())
	}

	pubKeyStr := C.GoString(ID)
	permissionStr := C.GoString(permission)

	callOk, RequireAuthCost := sbx.host.RequireAuth(pubKeyStr, permissionStr)

	*ok = C.bool(callOk)

	*gasUsed = C.size_t(RequireAuthCost.CPU)

	return nil
}

//export goReceipt
func goReceipt(cSbx C.SandboxPtr, content *C.char, gasUsed *C.size_t) *C.char {
	sbx, sbOk := GetSandbox(cSbx)
	if !sbOk {
		return C.CString(ErrGetSandbox.Error())
	}

	contentStr := C.GoString(content)

	cost := sbx.host.Receipt(contentStr)

	*gasUsed = C.size_t(cost.CPU)

	return nil
}

//export goEvent
func goEvent(cSbx C.SandboxPtr, content *C.char, gasUsed *C.size_t) *C.char {
	sbx, sbOk := GetSandbox(cSbx)
	if !sbOk {
		return C.CString(ErrGetSandbox.Error())
	}

	contentStr := C.GoString(content)

	cost := sbx.host.PostEvent(contentStr)

	*gasUsed = C.size_t(cost.CPU)

	return nil
}
