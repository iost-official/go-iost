package v8

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/Gaurav-Gosain/quickjs"
	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/core/contract"
	"github.com/iost-official/go-iost/v3/crypto"
	"github.com/iost-official/go-iost/v3/vm/database"
)

const cryptGasBase = 100

func getSbx(ctx *quickjs.Context) *Sandbox {
	v, ok := sbxMap.Load(ctx)
	if !ok {
		panic("get sandbox failed")
	}
	return v.(*Sandbox)
}

func newIOSTBlockchain(ctx *quickjs.Context) quickjs.Value {
	obj := ctx.Object()
	sbx := getSbx(ctx)

	obj.Set("blockInfo", ctx.Function("blockInfo", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		blkInfo, cost := sbx.host.BlockInfo()
		sbx.gasUsed += int64(cost.CPU)
		return ctx.String(string(blkInfo))
	}))
	obj.Set("txInfo", ctx.Function("txInfo", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		txInfo, cost := sbx.host.TxInfo()
		sbx.gasUsed += int64(cost.CPU)
		return ctx.String(string(txInfo))
	}))
	obj.Set("contextInfo", ctx.Function("contextInfo", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		ctxInfo, cost := sbx.host.ContextInfo()
		sbx.gasUsed += int64(cost.CPU)
		return ctx.String(string(ctxInfo))
	}))
	obj.Set("call", ctx.Function("call", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 3 {
			return ctx.ThrowError("IOSTBlockchain_call invalid argument length")
		}
		contract := args[0].String()
		api := args[1].String()
		jarg := args[2].String()
		callRs, cost, err := sbx.host.Call(contract, api, jarg)
		sbx.gasUsed += int64(cost.CPU)
		if err != nil {
			return ctx.ThrowError(err.Error())
		}
		rsStr, _ := json.Marshal(callRs)
		return ctx.String(string(rsStr))
	}))
	obj.Set("callWithAuth", ctx.Function("callWithAuth", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 3 {
			return ctx.ThrowError("IOSTBlockchain_callWithAuth invalid argument length")
		}
		contract := args[0].String()
		api := args[1].String()
		jarg := args[2].String()
		callRs, cost, err := sbx.host.CallWithAuth(contract, api, jarg)
		sbx.gasUsed += int64(cost.CPU)
		if err != nil {
			return ctx.ThrowError(err.Error())
		}
		rsStr, _ := json.Marshal(callRs)
		return ctx.String(string(rsStr))
	}))
	obj.Set("requireAuth", ctx.Function("requireAuth", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 2 {
			return ctx.ThrowError("IOSTBlockchain_requireAuth invalid argument length")
		}
		accountID := args[0].String()
		permission := args[1].String()
		ok, cost := sbx.host.RequireAuth(accountID, permission)
		sbx.gasUsed += int64(cost.CPU)
		return ctx.Bool(ok)
	}))
	obj.Set("receipt", ctx.Function("receipt", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 1 {
			return ctx.ThrowError("IOSTBlockchain_receipt invalid argument length")
		}
		content := args[0].String()
		cost := sbx.host.Receipt(content)
		sbx.gasUsed += int64(cost.CPU)
		return ctx.Null()
	}))
	obj.Set("event", ctx.Function("event", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 1 {
			return ctx.ThrowError("IOSTBlockchain_event invalid argument length")
		}
		content := args[0].String()
		cost := sbx.host.PostEvent(content)
		sbx.gasUsed += int64(cost.CPU)
		return ctx.Null()
	}))

	return obj
}

func newIOSTStorage(ctx *quickjs.Context) quickjs.Value {
	obj := ctx.Object()
	sbx := getSbx(ctx)

	obj.Set("put", ctx.Function("put", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 2 {
			return ctx.ThrowError("IOSTContractStorage_Put invalid argument length")
		}
		k := args[0].String()
		v := args[1].String()
		ramPayer := ""
		if len(args) > 2 {
			ramPayer = args[2].String()
		}
		var cost contract.Cost
		var err error
		if ramPayer == "" {
			cost, err = sbx.host.Put(k, v)
		} else {
			cost, err = sbx.host.Put(k, v, ramPayer)
		}
		sbx.gasUsed += int64(cost.CPU)
		if err != nil {
			return ctx.ThrowError(err.Error())
		}
		return ctx.Null()
	}))
	obj.Set("has", ctx.Function("has", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 1 {
			return ctx.ThrowError("IOSTContractStorage_Has invalid argument length")
		}
		k := args[0].String()
		ret, cost := sbx.host.Has(k)
		sbx.gasUsed += int64(cost.CPU)
		return ctx.Bool(ret)
	}))
	obj.Set("get", ctx.Function("get", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 1 {
			return ctx.ThrowError("IOSTContractStorage_Get invalid argument length")
		}
		k := args[0].String()
		val, cost := sbx.host.Get(k)
		sbx.gasUsed += int64(cost.CPU)
		if val == nil {
			return ctx.Null()
		}
		return ctx.String(dbValToString(val))
	}))
	obj.Set("del", ctx.Function("del", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 1 {
			return ctx.ThrowError("IOSTContractStorage_Del invalid argument length")
		}
		k := args[0].String()
		cost, err := sbx.host.Del(k)
		sbx.gasUsed += int64(cost.CPU)
		if err != nil {
			return ctx.ThrowError(err.Error())
		}
		return ctx.Null()
	}))

	obj.Set("mapPut", ctx.Function("mapPut", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 3 {
			return ctx.ThrowError("IOSTContractStorage_MapPut invalid argument length")
		}
		k := args[0].String()
		f := args[1].String()
		v := args[2].String()
		ramPayer := ""
		if len(args) > 3 {
			ramPayer = args[3].String()
		}
		var cost contract.Cost
		var err error
		if ramPayer == "" {
			cost, err = sbx.host.MapPut(k, f, v)
		} else {
			cost, err = sbx.host.MapPut(k, f, v, ramPayer)
		}
		sbx.gasUsed += int64(cost.CPU)
		if err != nil {
			return ctx.ThrowError(err.Error())
		}
		return ctx.Null()
	}))
	obj.Set("mapHas", ctx.Function("mapHas", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 2 {
			return ctx.ThrowError("IOSTContractStorage_MapHas invalid argument length")
		}
		k := args[0].String()
		f := args[1].String()
		ret, cost := sbx.host.MapHas(k, f)
		sbx.gasUsed += int64(cost.CPU)
		return ctx.Bool(ret)
	}))
	obj.Set("mapGet", ctx.Function("mapGet", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 2 {
			return ctx.ThrowError("IOSTContractStorage_MapGet invalid argument length")
		}
		k := args[0].String()
		f := args[1].String()
		val, cost := sbx.host.MapGet(k, f)
		sbx.gasUsed += int64(cost.CPU)
		if val == nil {
			return ctx.Null()
		}
		return ctx.String(dbValToString(val))
	}))
	obj.Set("mapDel", ctx.Function("mapDel", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 2 {
			return ctx.ThrowError("IOSTContractStorage_MapDel invalid argument length")
		}
		k := args[0].String()
		f := args[1].String()
		cost, err := sbx.host.MapDel(k, f)
		sbx.gasUsed += int64(cost.CPU)
		if err != nil {
			return ctx.ThrowError(err.Error())
		}
		return ctx.Null()
	}))
	obj.Set("mapKeys", ctx.Function("mapKeys", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 1 {
			return ctx.ThrowError("IOSTContractStorage_MapKeys invalid argument length")
		}
		k := args[0].String()
		fstr, cost := sbx.host.MapKeys(k)
		sbx.gasUsed += int64(cost.CPU)
		j, _ := json.Marshal(fstr)
		return ctx.String(string(j))
	}))
	obj.Set("mapLen", ctx.Function("mapLen", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 1 {
			return ctx.ThrowError("IOSTContractStorage_MapLen invalid argument length")
		}
		k := args[0].String()
		l, cost := sbx.host.MapLen(k)
		sbx.gasUsed += int64(cost.CPU)
		return ctx.Int32(int32(l))
	}))

	obj.Set("globalHas", ctx.Function("globalHas", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 2 {
			return ctx.ThrowError("IOSTContractStorage_GlobalHas invalid argument length")
		}
		c := args[0].String()
		k := args[1].String()
		ret, cost := sbx.host.GlobalHas(c, k)
		sbx.gasUsed += int64(cost.CPU)
		return ctx.Bool(ret)
	}))
	obj.Set("globalGet", ctx.Function("globalGet", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 2 {
			return ctx.ThrowError("IOSTContractStorage_GlobalGet invalid argument length")
		}
		c := args[0].String()
		k := args[1].String()
		val, cost := sbx.host.GlobalGet(c, k)
		sbx.gasUsed += int64(cost.CPU)
		if val == nil {
			return ctx.Null()
		}
		return ctx.String(dbValToString(val))
	}))
	obj.Set("globalMapHas", ctx.Function("globalMapHas", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 3 {
			return ctx.ThrowError("IOSTContractStorage_GlobalMapHas invalid argument length")
		}
		c := args[0].String()
		k := args[1].String()
		f := args[2].String()
		ret, cost := sbx.host.GlobalMapHas(c, k, f)
		sbx.gasUsed += int64(cost.CPU)
		return ctx.Bool(ret)
	}))
	obj.Set("globalMapGet", ctx.Function("globalMapGet", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 3 {
			return ctx.ThrowError("IOSTContractStorage_GlobalMapGet invalid argument length")
		}
		c := args[0].String()
		k := args[1].String()
		f := args[2].String()
		val, cost := sbx.host.GlobalMapGet(c, k, f)
		sbx.gasUsed += int64(cost.CPU)
		if val == nil {
			return ctx.Null()
		}
		return ctx.String(dbValToString(val))
	}))
	obj.Set("globalMapKeys", ctx.Function("globalMapKeys", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 2 {
			return ctx.ThrowError("IOSTContractStorage_GlobalMapKeys invalid argument length")
		}
		c := args[0].String()
		k := args[1].String()
		fstr, cost := sbx.host.GlobalMapKeys(c, k)
		sbx.gasUsed += int64(cost.CPU)
		j, _ := json.Marshal(fstr)
		return ctx.String(string(j))
	}))
	obj.Set("globalMapLen", ctx.Function("globalMapLen", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 2 {
			return ctx.ThrowError("IOSTContractStorage_GlobalMapLen invalid argument length")
		}
		c := args[0].String()
		k := args[1].String()
		l, cost := sbx.host.GlobalMapLen(c, k)
		sbx.gasUsed += int64(cost.CPU)
		return ctx.Int32(int32(l))
	}))

	return obj
}

func newIOSTInstruction(ctx *quickjs.Context) quickjs.Value {
	obj := ctx.Object()
	sbx := getSbx(ctx)

	obj.Set("incr", ctx.Function("incr", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 1 {
			return ctx.ThrowError("IOSTContractInstruction_Incr invalid argument length")
		}
		n, err := args[0].Int32()
		if err != nil {
			return ctx.ThrowError("IOSTContractInstruction_Incr value must be number")
		}
		if n < 0 {
			return ctx.ThrowError("IOSTContractInstruction_Incr invalid gas")
		}
		sbx.gasUsed += int64(n)
		return ctx.Int32(int32(sbx.gasUsed))
	}))
	obj.Set("count", ctx.Function("count", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		return ctx.Int32(int32(sbx.gasUsed))
	}))

	return obj
}

func newIOSTCrypto(ctx *quickjs.Context) quickjs.Value {
	obj := ctx.Object()

	obj.Set("sha3", ctx.Function("sha3", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 1 {
			return ctx.ThrowError("IOSTCrypto_sha3 invalid argument length")
		}
		msg := args[0].String()
		val := common.Base58Encode(common.Sha3([]byte(msg)))
		sbx := getSbx(ctx)
		sbx.gasUsed += int64(len(msg) + cryptGasBase)
		return ctx.String(val)
	}))
	obj.Set("sha3Hex", ctx.Function("sha3Hex", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 1 {
			return ctx.ThrowError("IOSTCrypto_sha3Hex invalid argument length")
		}
		msg := args[0].String()
		msgBytes, err := hex.DecodeString(msg)
		sbx := getSbx(ctx)
		sbx.gasUsed += int64(len(msgBytes) + cryptGasBase)
		if err != nil {
			return ctx.String("")
		}
		val := hex.EncodeToString(common.Sha3(msgBytes))
		return ctx.String(val)
	}))
	obj.Set("ripemd160Hex", ctx.Function("ripemd160Hex", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 1 {
			return ctx.ThrowError("IOSTCrypto_ripemd160Hex invalid argument length")
		}
		msg := args[0].String()
		msgBytes, err := hex.DecodeString(msg)
		sbx := getSbx(ctx)
		sbx.gasUsed += int64(len(msgBytes) + cryptGasBase)
		if err != nil {
			return ctx.String("")
		}
		val := hex.EncodeToString(common.Ripemd160(msgBytes))
		return ctx.String(val)
	}))
	obj.Set("verify", ctx.Function("verify", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 4 {
			return ctx.ThrowError("IOSTCrypto_verify invalid argument length")
		}
		algoStr := args[0].String()
		msgBytes := common.Base58Decode(args[1].String())
		sigBytes := common.Base58Decode(args[2].String())
		pubkeyBytes := common.Base58Decode(args[3].String())
		sbx := getSbx(ctx)
		sbx.gasUsed += int64(len(msgBytes) + cryptGasBase)
		if algoStr != "secp256k1" && algoStr != "ed25519" {
			return ctx.Int32(0)
		}
		if !crypto.NewAlgorithm(algoStr).Verify(msgBytes, pubkeyBytes, sigBytes) {
			return ctx.Int32(0)
		}
		return ctx.Int32(1)
	}))

	return obj
}

func newConsole(ctx *quickjs.Context) quickjs.Value {
	obj := ctx.Object()
	sbx := getSbx(ctx)

	obj.Set("log", ctx.Function("log", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 2 {
			return ctx.ThrowError("console log invalid argument length")
		}
		levelStr := args[0].String()
		detailStr := args[1].String()

		if sbx.host.Logger() == nil {
			return ctx.ThrowError("no logger error")
		}

		loggerVal := reflect.ValueOf(sbx.host.Logger())
		loggerFunc := loggerVal.MethodByName(levelStr)
		if !loggerFunc.IsValid() {
			return ctx.ThrowError("log invalid level")
		}

		loggerFunc.Call([]reflect.Value{
			reflect.ValueOf(detailStr),
		})
		return ctx.Undefined()
	}))

	return obj
}

func newCLog(ctx *quickjs.Context) quickjs.Value {
	return ctx.Function("_cLog", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) < 2 {
			return ctx.Undefined()
		}
		levelStr := args[0].String()
		detailStr := args[1].String()

		sbx := getSbx(ctx)
		if sbx.host == nil || sbx.host.Logger() == nil {
			fmt.Printf("[JSLOG %s] %s\n", levelStr, detailStr)
			return ctx.Undefined()
		}

		loggerVal := reflect.ValueOf(sbx.host.Logger())
		loggerFunc := loggerVal.MethodByName(levelStr)
		if !loggerFunc.IsValid() {
			return ctx.Undefined()
		}

		loggerFunc.Call([]reflect.Value{
			reflect.ValueOf(detailStr),
		})
		return ctx.Undefined()
	})
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
