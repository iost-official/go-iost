package v8

/*
#include "v8/vm.h"
*/
import "C"
import "encoding/json"

//export goTransfer
func goTransfer(cSbx C.SandboxPtr, from, to, amount *C.char, gasUsed *C.size_t) C.int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	fromStr := C.GoString(from)
	toStr := C.GoString(to)
	//amountStr := C.GoString(amount)

	ret := C.int(0)
	_, err := sbx.host.Transfer(fromStr, toStr, 0)
	if err != nil {
		ret = C.int(1)
	}
	return ret
}

//export goWithdraw
func goWithdraw(cSbx C.SandboxPtr, to, amount *C.char, gasUsed *C.size_t) C.int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	toStr := C.GoString(to)
	//amountStr := C.GoString(amount)

	ret := C.int(0)
	_, err := sbx.host.Withdraw(toStr, 0)
	if err != nil {
		ret = C.int(1)
	}
	return ret
}

//export goDeposit
func goDeposit(cSbx C.SandboxPtr, from, amount *C.char, gasUsed *C.size_t) C.int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	fromStr := C.GoString(from)
	//amountStr := C.GoString(amount)

	ret := C.int(0)
	_, err := sbx.host.Deposit(fromStr, 0)
	if err != nil {
		ret = C.int(1)
	}
	return ret
}

//export goTopUp
func goTopUp(cSbx C.SandboxPtr, contract, from, amount *C.char, gasUsed *C.size_t) C.int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	contractStr := C.GoString(contract)
	fromStr := C.GoString(from)
	//amountStr := C.GoString(amount)

	ret := C.int(0)
	_, err := sbx.host.TopUp(contractStr, fromStr, 0)
	if err != nil {
		ret = C.int(1)
	}
	return ret
}

//export goCountermand
func goCountermand(cSbx C.SandboxPtr, contract, to, amount *C.char, gasUsed *C.size_t) C.int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	contractStr := C.GoString(contract)
	toStr := C.GoString(to)
	//amountStr := C.GoString(amount)

	ret := C.int(0)
	_, err := sbx.host.TopUp(contractStr, toStr, 0)
	if err != nil {
		ret = C.int(1)
	}
	return ret
}

//export goBlockInfo
func goBlockInfo(cSbx C.SandboxPtr, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	blkInfo, _ := sbx.host.BlockInfo()
	return C.CString(string(blkInfo))
}

//export goTxInfo
func goTxInfo(cSbx C.SandboxPtr, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	txInfo, _ := sbx.host.TxInfo()
	return C.CString(string(txInfo))
}

//export goCall
func goCall(cSbx C.SandboxPtr, contract, api, args *C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	contractStr := C.GoString(contract)
	apiStr := C.GoString(api)
	argsStr := C.GoString(args)

	callRs, _, _ := sbx.host.Call(contractStr, apiStr, argsStr)

	rsStr, _ := json.Marshal(callRs)

	return C.CString(string(rsStr))
}
