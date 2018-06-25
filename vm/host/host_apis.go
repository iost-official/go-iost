/*
Define host api
*/
package host

import (
	"errors"

	"os"

	"fmt"
	"time"

	"encoding/json"

	"strconv"

	"github.com/iost-official/gopher-lua"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/log"
	"github.com/iost-official/prototype/vm"
)

var l log.Logger

var logFile *os.File

func init() {
	var err error
	logFile, err = os.OpenFile("vm.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
}

var (
	ErrBalanceNotEnough = errors.New("balance not enough")
)

func Put(pool state.Pool, key state.Key, value state.Value) bool {
	pool.Put(key, value)
	return true
}

func Get(pool state.Pool, key state.Key) (state.Value, error) {
	return pool.Get(key)
}

func Log(s, cid string) {
	str := fmt.Sprintf("%v %v/%v: %v", time.Now().Format("2006-01-02 15:04:05.000"), "lua", cid, s)
	logFile.Write([]byte(str))
	logFile.Write([]byte("\n"))
}

func Transfer(pool state.Pool, src, des string, value float64) bool {

	err := changeToken(pool, "iost", state.Key(src), -value)

	if err != nil {
		return false
	}

	err = changeToken(pool, "iost", state.Key(des), value)
	if err != nil {
		return false
	}

	return true
}

func Deposit(pool state.Pool, contractPrefix, payer string, value float64) bool {
	err := changeToken(pool, "iost", state.Key(payer), -value)
	if err != nil {
		return false
	}

	err = changeToken(pool, "iost-contract", state.Key(contractPrefix), value)
	if err != nil {
		return false
	}

	return true

}

func Withdraw(pool state.Pool, contractPrefix, payer string, value float64) bool {
	err := changeToken(pool, "iost-contract", state.Key(contractPrefix), -value)
	if err != nil {
		return false
	}
	err = changeToken(pool, "iost", state.Key(payer), value)
	if err != nil {
		return false
	}

	return true

}
func RandomByParentHash(ctx *vm.Context, probability float64) bool {
	var seed []byte
	for ctx != nil {
		if ctx.ParentHash == nil {
			ctx = ctx.Base
		} else {
			seed = ctx.ParentHash
			break
		}
	}
	seed = ctx.ParentHash

	return float64(common.Sha256(seed)[10]) > probability*255
}

func ParentHashLast(ctx *vm.Context) byte {
	for ctx != nil {
		if ctx.ParentHash == nil {
			ctx = ctx.Base
		} else {
			return ctx.ParentHash[len(ctx.ParentHash)-1]
		}
	}
	return 0
}

func Publisher(contract vm.Contract) string {
	return string(contract.Info().Publisher)
}

func Now(ctx *vm.Context) int64 {
	for ctx != nil {
		if ctx.Timestamp <= int64(0) {
			ctx = ctx.Base
		} else {
			return ctx.Timestamp
		}
	}
	return 0
}

func BlockHeight(ctx *vm.Context) int64 {
	for ctx != nil {
		if ctx.BlockHeight <= int64(0) {
			ctx = ctx.Base
		} else {
			return ctx.BlockHeight
		}
	}
	return 0
}

func Witness(ctx *vm.Context) string {
	for ctx != nil {
		if len(ctx.Witness) <= 0 {
			ctx = ctx.Base
		} else {
			return string(ctx.Witness)
		}
	}
	return ""
}

func TableToJson(table *lua.LTable) (string, error) { // todo let state.pool works well
	m := tableIterator(table)

	rtn, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(rtn), nil
}

func tableIterator(table *lua.LTable) map[string]interface{} {
	m := make(map[string]interface{})
	table.ForEach(func(key lua.LValue, val lua.LValue) {
		var k0 string
		if key.Type() == lua.LTNumber {
			k0 = strconv.Itoa(int(float64(key.(lua.LNumber))))
		} else {
			k0 = key.String()
		}

		switch val.(type) {
		case lua.LNumber:
			f := float64(val.(lua.LNumber))
			m[k0] = f
		case lua.LString:
			s := val.String()
			m[k0] = s
		case lua.LBool:
			m[k0] = val == lua.LTrue
		case *lua.LTable:
			m[k0] = tableIterator(val.(*lua.LTable))
		}
	})
	return m
}

func ParseJson(jsonStr []byte) (*lua.LTable, error) {
	var mapResult map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &mapResult); err != nil {
		return nil, err
	}
	//fmt.Println("in ParseJson:", mapResult)

	return mapIterator(mapResult), nil
}

func mapIterator(m map[string]interface{}) *lua.LTable {
	lt := lua.LTable{}

	for k, v := range m {
		var l lua.LValue
		switch v.(type) {
		case float64:
			l = lua.LNumber(v.(float64))
		case string:
			l = lua.LString(v.(string))
		case bool:
			l = lua.LBool(v.(bool))
		case map[string]interface{}:
			l = mapIterator(v.(map[string]interface{}))
		}
		//fmt.Println(k, l)
		setTable(&lt, k, l)
	}
	//lt.ForEach(func(value lua.LValue, value2 lua.LValue) {
	//	fmt.Println("in lt:", value, value2)
	//})
	return &lt
}

func setTable(lt *lua.LTable, k string, v lua.LValue) {
	i, err := strconv.Atoi(k)
	if err == nil {
		lt.RawSetInt(i, v)
	} else {
		lt.RawSetString(k, v)
	}
}

func changeToken(pool state.Pool, key, field state.Key, delta float64) error {
	val0, err := pool.GetHM(state.Key(key), state.Key(field))
	if err != nil {
		return err
	}
	var val float64
	switch val0.(type) {
	case *state.VFloat:
		val = val0.(*state.VFloat).ToFloat64()
	default:
		val = 0
	}

	if val+delta < 0 {
		return ErrBalanceNotEnough
	}
	ba := state.MakeVFloat(val + delta)

	pool.PutHM(state.Key(key), state.Key(field), ba)
	return nil
}
