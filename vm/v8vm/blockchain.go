package v8

/*
#include "v8/vm.h"
*/
import "C"
import (
	"encoding/json"
	"strconv"

	"github.com/iost-official/go-iost/vm/host"
)

// todo replace this error code with c++ error code
// transfer err list
const (
	TransferSuccess = iota
	TransferInvalidAmount
	TransferBalanceNotEnough
	TransferUnexpectedError
)

// blockInfo err list
const (
	BlockInfoSuccess = iota
	BlockInfoUnexpectedError
)

// txInfo err list
const (
	TxInfoSuccess = iota
	TxInfoUnexpectedError
)

// contractCall err list
const (
	ContractCallSuccess = iota
	ContractCallUnexpectedError
)

// APICall err list
const (
	APICallSuccess = iota
	APICallUnexpectedError
)

//export goTransfer
func goTransfer(cSbx C.SandboxPtr, from, to, amount *C.char, gasUsed *C.size_t) int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return TransferUnexpectedError
	}

	fromStr := C.GoString(from)
	toStr := C.GoString(to)
	amountStr := C.GoString(amount)
	amountInt64, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil || amountInt64 <= 0 {
		return TransferInvalidAmount
	}

	cost, err := sbx.host.Transfer(fromStr, toStr, amountInt64)
	*gasUsed = C.size_t(cost.Data)

	if err != nil && err == host.ErrBalanceNotEnough {
		return TransferBalanceNotEnough
	}

	return TransferSuccess
}

//export goWithdraw
func goWithdraw(cSbx C.SandboxPtr, to, amount *C.char, gasUsed *C.size_t) int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return TransferUnexpectedError
	}

	toStr := C.GoString(to)
	amountStr := C.GoString(amount)
	amountInt64, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil || amountInt64 <= 0 {
		return TransferInvalidAmount
	}

	cost, err := sbx.host.Withdraw(toStr, amountInt64)
	*gasUsed = C.size_t(cost.Data)

	if err != nil && err == host.ErrBalanceNotEnough {
		return TransferBalanceNotEnough
	}

	return TransferSuccess
}

//export goDeposit
func goDeposit(cSbx C.SandboxPtr, from, amount *C.char, gasUsed *C.size_t) int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return TransferUnexpectedError
	}

	fromStr := C.GoString(from)
	amountStr := C.GoString(amount)
	amountInt64, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil || amountInt64 <= 0 {
		return TransferInvalidAmount
	}

	cost, err := sbx.host.Deposit(fromStr, amountInt64)
	*gasUsed = C.size_t(cost.Data)

	if err != nil && err == host.ErrBalanceNotEnough {
		return TransferBalanceNotEnough
	}

	return TransferSuccess
}

//export goTopUp
func goTopUp(cSbx C.SandboxPtr, contract, from, amount *C.char, gasUsed *C.size_t) int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return TransferUnexpectedError
	}

	contractStr := C.GoString(contract)
	fromStr := C.GoString(from)
	amountStr := C.GoString(amount)
	amountInt64, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil || amountInt64 <= 0 {
		return TransferInvalidAmount
	}

	cost, err := sbx.host.TopUp(contractStr, fromStr, amountInt64)
	*gasUsed = C.size_t(cost.Data)

	if err != nil && err == host.ErrBalanceNotEnough {
		return TransferBalanceNotEnough
	}

	return TransferSuccess
}

//export goCountermand
func goCountermand(cSbx C.SandboxPtr, contract, to, amount *C.char, gasUsed *C.size_t) int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return TransferUnexpectedError
	}

	contractStr := C.GoString(contract)
	toStr := C.GoString(to)
	amountStr := C.GoString(amount)
	amountInt64, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil || amountInt64 <= 0 {
		return TransferInvalidAmount
	}

	cost, err := sbx.host.Countermand(contractStr, toStr, amountInt64)
	*gasUsed = C.size_t(cost.Data)

	if err != nil && err == host.ErrBalanceNotEnough {
		return TransferBalanceNotEnough
	}

	return TransferSuccess
}

//export goBlockInfo
func goBlockInfo(cSbx C.SandboxPtr, info **C.char, gasUsed *C.size_t) int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return BlockInfoUnexpectedError
	}

	blkInfo, cost := sbx.host.BlockInfo()
	*gasUsed = C.size_t(cost.Data)
	*info = C.CString(string(blkInfo))

	return BlockInfoSuccess
}

//export goTxInfo
func goTxInfo(cSbx C.SandboxPtr, info **C.char, gasUsed *C.size_t) int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return TxInfoUnexpectedError
	}

	txInfo, cost := sbx.host.TxInfo()
	*gasUsed = C.size_t(cost.Data)
	*info = C.CString(string(txInfo))

	return TxInfoSuccess
}

//export goCall
func goCall(cSbx C.SandboxPtr, contract, api, args *C.char, result **C.char, gasUsed *C.size_t) int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return ContractCallUnexpectedError
	}

	contractStr := C.GoString(contract)
	apiStr := C.GoString(api)
	argsStr := C.GoString(args)

	callRs, cost, err := sbx.host.Call(contractStr, apiStr, argsStr)
	*gasUsed = C.size_t(cost.Data)
	if err != nil {
		return ContractCallUnexpectedError
	}

	rsStr, err := json.Marshal(callRs)
	if err != nil {
		return ContractCallUnexpectedError
	}

	*result = C.CString(string(rsStr))

	return ContractCallSuccess
}

//export goCallWithReceipt
func goCallWithReceipt(cSbx C.SandboxPtr, contract, api, args *C.char, result **C.char, gasUsed *C.size_t) int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return ContractCallUnexpectedError
	}

	contractStr := C.GoString(contract)
	apiStr := C.GoString(api)
	argsStr := C.GoString(args)

	callRs, cost, err := sbx.host.CallWithReceipt(contractStr, apiStr, argsStr)
	*gasUsed = C.size_t(cost.Data)
	if err != nil {
		return ContractCallUnexpectedError
	}

	rsStr, err := json.Marshal(callRs)
	if err != nil {
		return ContractCallUnexpectedError
	}

	*result = C.CString(string(rsStr))

	return ContractCallSuccess
}

//export goRequireAuth
func goRequireAuth(cSbx C.SandboxPtr, pubKey *C.char, ok *C.bool, gasUsed *C.size_t) int {
	sbx, sbOk := GetSandbox(cSbx)
	if !sbOk {
		return APICallUnexpectedError
	}

	pubKeyStr := C.GoString(pubKey)

	callOk, RequireAuthCost := sbx.host.APIDelegate.RequireAuth(pubKeyStr)

	*ok = C.bool(callOk)
	if callOk != true {
		return APICallUnexpectedError
	}

	*gasUsed = C.size_t(RequireAuthCost.Data)

	return APICallSuccess
}

//export goGrantServi
func goGrantServi(cSbx C.SandboxPtr, pubKey *C.char, amount *C.char, gasUsed *C.size_t) int {
	sbx, sbOk := GetSandbox(cSbx)
	if !sbOk {
		return APICallUnexpectedError
	}

	pubKeyStr := C.GoString(pubKey)
	amountStr := C.GoString(amount)
	amountInt64, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil || amountInt64 <= 0 {
		return APICallUnexpectedError
	}

	cost, err := sbx.host.GrantServi(pubKeyStr, amountInt64)
	*gasUsed = C.size_t(cost.Data)

	if err != nil {
		return APICallUnexpectedError
	}

	return APICallSuccess
}
