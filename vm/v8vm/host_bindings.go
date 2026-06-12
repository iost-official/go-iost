package v8

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/dop251/goja"
	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/core/contract"
	"github.com/iost-official/go-iost/v3/crypto"
	"github.com/iost-official/go-iost/v3/vm/database"
)

const cryptGasBase = 100

func newIOSTBlockchain(sbx *Sandbox) *goja.Object {
	obj := sbx.rt.NewObject()

	obj.Set("blockInfo", func(call goja.FunctionCall) goja.Value {
		blkInfo, cost := sbx.host.BlockInfo()
		sbx.gasUsed += cost.CPU
		return sbx.rt.ToValue(string(blkInfo))
	})
	obj.Set("txInfo", func(call goja.FunctionCall) goja.Value {
		txInfo, cost := sbx.host.TxInfo()
		sbx.gasUsed += cost.CPU
		return sbx.rt.ToValue(string(txInfo))
	})
	obj.Set("contextInfo", func(call goja.FunctionCall) goja.Value {
		ctxInfo, cost := sbx.host.ContextInfo()
		sbx.gasUsed += cost.CPU
		return sbx.rt.ToValue(string(ctxInfo))
	})
	obj.Set("call", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			panic(sbx.rt.NewTypeError("IOSTBlockchain_call invalid argument length"))
		}
		contract := call.Argument(0).String()
		api := call.Argument(1).String()
		jarg := call.Argument(2).String()
		callRs, cost, err := sbx.host.Call(contract, api, jarg)
		sbx.gasUsed += cost.CPU
		if err != nil {
			panic(sbx.rt.NewTypeError(err.Error()))
		}
		rsStr, _ := json.Marshal(callRs)
		return sbx.rt.ToValue(string(rsStr))
	})
	obj.Set("callWithAuth", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			panic(sbx.rt.NewTypeError("IOSTBlockchain_callWithAuth invalid argument length"))
		}
		contract := call.Argument(0).String()
		api := call.Argument(1).String()
		jarg := call.Argument(2).String()
		callRs, cost, err := sbx.host.CallWithAuth(contract, api, jarg)
		sbx.gasUsed += cost.CPU
		if err != nil {
			panic(sbx.rt.NewTypeError(err.Error()))
		}
		rsStr, _ := json.Marshal(callRs)
		return sbx.rt.ToValue(string(rsStr))
	})
	obj.Set("requireAuth", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(sbx.rt.NewTypeError("IOSTBlockchain_requireAuth invalid argument length"))
		}
		accountID := call.Argument(0).String()
		permission := call.Argument(1).String()
		ok, cost := sbx.host.RequireAuth(accountID, permission)
		sbx.gasUsed += cost.CPU
		return sbx.rt.ToValue(ok)
	})
	obj.Set("receipt", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(sbx.rt.NewTypeError("IOSTBlockchain_receipt invalid argument length"))
		}
		content := call.Argument(0).String()
		cost := sbx.host.Receipt(content)
		sbx.gasUsed += cost.CPU
		return goja.Null()
	})
	obj.Set("event", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(sbx.rt.NewTypeError("IOSTBlockchain_event invalid argument length"))
		}
		content := call.Argument(0).String()
		cost := sbx.host.PostEvent(content)
		sbx.gasUsed += cost.CPU
		return goja.Null()
	})

	return obj
}

func newIOSTStorage(sbx *Sandbox) *goja.Object {
	obj := sbx.rt.NewObject()

	obj.Set("put", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(sbx.rt.NewTypeError("IOSTContractStorage_Put invalid argument length"))
		}
		k := call.Argument(0).String()
		v := call.Argument(1).String()
		ramPayer := ""
		if len(call.Arguments) > 2 {
			ramPayer = call.Argument(2).String()
		}
		var cost contract.Cost
		var err error
		if ramPayer == "" {
			cost, err = sbx.host.Put(k, v)
		} else {
			cost, err = sbx.host.Put(k, v, ramPayer)
		}
		sbx.gasUsed += cost.CPU
		if err != nil {
			panic(sbx.rt.NewTypeError(err.Error()))
		}
		return goja.Null()
	})
	obj.Set("has", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(sbx.rt.NewTypeError("IOSTContractStorage_Has invalid argument length"))
		}
		k := call.Argument(0).String()
		ret, cost := sbx.host.Has(k)
		sbx.gasUsed += cost.CPU
		return sbx.rt.ToValue(ret)
	})
	obj.Set("get", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(sbx.rt.NewTypeError("IOSTContractStorage_Get invalid argument length"))
		}
		k := call.Argument(0).String()
		val, cost := sbx.host.Get(k)
		sbx.gasUsed += cost.CPU
		if val == nil {
			return goja.Null()
		}
		return sbx.rt.ToValue(dbValToString(val))
	})
	obj.Set("del", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(sbx.rt.NewTypeError("IOSTContractStorage_Del invalid argument length"))
		}
		k := call.Argument(0).String()
		cost, err := sbx.host.Del(k)
		sbx.gasUsed += cost.CPU
		if err != nil {
			panic(sbx.rt.NewTypeError(err.Error()))
		}
		return goja.Null()
	})

	obj.Set("mapPut", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			panic(sbx.rt.NewTypeError("IOSTContractStorage_MapPut invalid argument length"))
		}
		k := call.Argument(0).String()
		f := call.Argument(1).String()
		v := call.Argument(2).String()
		ramPayer := ""
		if len(call.Arguments) > 3 {
			ramPayer = call.Argument(3).String()
		}
		var cost contract.Cost
		var err error
		if ramPayer == "" {
			cost, err = sbx.host.MapPut(k, f, v)
		} else {
			cost, err = sbx.host.MapPut(k, f, v, ramPayer)
		}
		sbx.gasUsed += cost.CPU
		if err != nil {
			panic(sbx.rt.NewTypeError(err.Error()))
		}
		return goja.Null()
	})
	obj.Set("mapHas", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(sbx.rt.NewTypeError("IOSTContractStorage_MapHas invalid argument length"))
		}
		k := call.Argument(0).String()
		f := call.Argument(1).String()
		ret, cost := sbx.host.MapHas(k, f)
		sbx.gasUsed += cost.CPU
		return sbx.rt.ToValue(ret)
	})
	obj.Set("mapGet", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(sbx.rt.NewTypeError("IOSTContractStorage_MapGet invalid argument length"))
		}
		k := call.Argument(0).String()
		f := call.Argument(1).String()
		val, cost := sbx.host.MapGet(k, f)
		sbx.gasUsed += cost.CPU
		if val == nil {
			return goja.Null()
		}
		return sbx.rt.ToValue(dbValToString(val))
	})
	obj.Set("mapDel", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(sbx.rt.NewTypeError("IOSTContractStorage_MapDel invalid argument length"))
		}
		k := call.Argument(0).String()
		f := call.Argument(1).String()
		cost, err := sbx.host.MapDel(k, f)
		sbx.gasUsed += cost.CPU
		if err != nil {
			panic(sbx.rt.NewTypeError(err.Error()))
		}
		return goja.Null()
	})
	obj.Set("mapKeys", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(sbx.rt.NewTypeError("IOSTContractStorage_MapKeys invalid argument length"))
		}
		k := call.Argument(0).String()
		fstr, cost := sbx.host.MapKeys(k)
		sbx.gasUsed += cost.CPU
		j, _ := json.Marshal(fstr)
		return sbx.rt.ToValue(string(j))
	})
	obj.Set("mapLen", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(sbx.rt.NewTypeError("IOSTContractStorage_MapLen invalid argument length"))
		}
		k := call.Argument(0).String()
		l, cost := sbx.host.MapLen(k)
		sbx.gasUsed += cost.CPU
		return sbx.rt.ToValue(int32(l))
	})

	obj.Set("globalHas", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(sbx.rt.NewTypeError("IOSTContractStorage_GlobalHas invalid argument length"))
		}
		c := call.Argument(0).String()
		k := call.Argument(1).String()
		ret, cost := sbx.host.GlobalHas(c, k)
		sbx.gasUsed += cost.CPU
		return sbx.rt.ToValue(ret)
	})
	obj.Set("globalGet", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(sbx.rt.NewTypeError("IOSTContractStorage_GlobalGet invalid argument length"))
		}
		c := call.Argument(0).String()
		k := call.Argument(1).String()
		val, cost := sbx.host.GlobalGet(c, k)
		sbx.gasUsed += cost.CPU
		if val == nil {
			return goja.Null()
		}
		return sbx.rt.ToValue(dbValToString(val))
	})
	obj.Set("globalMapHas", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			panic(sbx.rt.NewTypeError("IOSTContractStorage_GlobalMapHas invalid argument length"))
		}
		c := call.Argument(0).String()
		k := call.Argument(1).String()
		f := call.Argument(2).String()
		ret, cost := sbx.host.GlobalMapHas(c, k, f)
		sbx.gasUsed += cost.CPU
		return sbx.rt.ToValue(ret)
	})
	obj.Set("globalMapGet", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			panic(sbx.rt.NewTypeError("IOSTContractStorage_GlobalMapGet invalid argument length"))
		}
		c := call.Argument(0).String()
		k := call.Argument(1).String()
		f := call.Argument(2).String()
		val, cost := sbx.host.GlobalMapGet(c, k, f)
		sbx.gasUsed += cost.CPU
		if val == nil {
			return goja.Null()
		}
		return sbx.rt.ToValue(dbValToString(val))
	})
	obj.Set("globalMapKeys", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(sbx.rt.NewTypeError("IOSTContractStorage_GlobalMapKeys invalid argument length"))
		}
		c := call.Argument(0).String()
		k := call.Argument(1).String()
		fstr, cost := sbx.host.GlobalMapKeys(c, k)
		sbx.gasUsed += cost.CPU
		j, _ := json.Marshal(fstr)
		return sbx.rt.ToValue(string(j))
	})
	obj.Set("globalMapLen", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(sbx.rt.NewTypeError("IOSTContractStorage_GlobalMapLen invalid argument length"))
		}
		c := call.Argument(0).String()
		k := call.Argument(1).String()
		l, cost := sbx.host.GlobalMapLen(c, k)
		sbx.gasUsed += cost.CPU
		return sbx.rt.ToValue(int32(l))
	})

	return obj
}

func newIOSTInstruction(sbx *Sandbox) *goja.Object {
	obj := sbx.rt.NewObject()

	obj.Set("incr", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(sbx.rt.NewTypeError("IOSTContractInstruction_Incr invalid argument length"))
		}
		n := call.Argument(0).ToInteger()
		if n < 0 {
			panic(sbx.rt.NewTypeError("IOSTContractInstruction_Incr invalid gas"))
		}
		sbx.gasUsed += n
		if sbx.gasUsed > sbx.gasLimit {
			panic(sbx.rt.NewTypeError("out of gas"))
		}
		return sbx.rt.ToValue(int32(sbx.gasUsed))
	})
	obj.Set("count", func(call goja.FunctionCall) goja.Value {
		return sbx.rt.ToValue(int32(sbx.gasUsed))
	})

	return obj
}

func newIOSTCrypto(sbx *Sandbox) *goja.Object {
	obj := sbx.rt.NewObject()

	obj.Set("sha3", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(sbx.rt.NewTypeError("IOSTCrypto_sha3 invalid argument length"))
		}
		msg := call.Argument(0).String()
		val := common.Base58Encode(common.Sha3([]byte(msg)))
		sbx.gasUsed += int64(len(msg) + cryptGasBase)
		return sbx.rt.ToValue(val)
	})
	obj.Set("sha3Hex", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(sbx.rt.NewTypeError("IOSTCrypto_sha3Hex invalid argument length"))
		}
		msg := call.Argument(0).String()
		msgBytes, err := hex.DecodeString(msg)
		sbx.gasUsed += int64(len(msgBytes) + cryptGasBase)
		if err != nil {
			return sbx.rt.ToValue("")
		}
		val := hex.EncodeToString(common.Sha3(msgBytes))
		return sbx.rt.ToValue(val)
	})
	obj.Set("ripemd160Hex", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(sbx.rt.NewTypeError("IOSTCrypto_ripemd160Hex invalid argument length"))
		}
		msg := call.Argument(0).String()
		msgBytes, err := hex.DecodeString(msg)
		sbx.gasUsed += int64(len(msgBytes) + cryptGasBase)
		if err != nil {
			return sbx.rt.ToValue("")
		}
		val := hex.EncodeToString(common.Ripemd160(msgBytes))
		return sbx.rt.ToValue(val)
	})
	obj.Set("verify", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 4 {
			panic(sbx.rt.NewTypeError("IOSTCrypto_verify invalid argument length"))
		}
		algoStr := call.Argument(0).String()
		msgBytes := common.Base58Decode(call.Argument(1).String())
		sigBytes := common.Base58Decode(call.Argument(2).String())
		pubkeyBytes := common.Base58Decode(call.Argument(3).String())
		sbx.gasUsed += int64(len(msgBytes) + cryptGasBase)
		if algoStr != "secp256k1" && algoStr != "ed25519" {
			return sbx.rt.ToValue(int32(0))
		}
		if !crypto.NewAlgorithm(algoStr).Verify(msgBytes, pubkeyBytes, sigBytes) {
			return sbx.rt.ToValue(int32(0))
		}
		return sbx.rt.ToValue(int32(1))
	})

	return obj
}

func newCLog(sbx *Sandbox) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		levelStr := call.Argument(0).String()
		detailStr := call.Argument(1).String()

		if sbx.host == nil || sbx.host.Logger() == nil {
			fmt.Printf("[JSLOG %s] %s\n", levelStr, detailStr)
			return goja.Undefined()
		}

		loggerVal := reflect.ValueOf(sbx.host.Logger())
		loggerFunc := loggerVal.MethodByName(levelStr)
		if !loggerFunc.IsValid() {
			return goja.Undefined()
		}

		loggerFunc.Call([]reflect.Value{
			reflect.ValueOf(detailStr),
		})
		return goja.Undefined()
	}
}

func dbValToString(val any) string {
	switch v := val.(type) {
	case int64:
		return strconv.FormatInt(v, 10)
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case []byte:
		return string(v)
	case database.SerializedJSON:
		return string(v)
	default:
		return ""
	}
}
