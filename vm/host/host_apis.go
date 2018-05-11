package host

import (
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/log"
)

var l log.Logger

func Put(pool state.Pool, key state.Key, value state.Value) bool {
	pool.Put(key, value)
	return true
}

func Get(pool state.Pool, key state.Key) (state.Value, error) {
	return pool.Get(key)
}

func Log(s, cid string) {
	l.D("From Lua %v > %v", cid, s)
}

func Transfer(pool state.Pool, src, des string, value float64) bool {
	val0, err := pool.GetHM("iost", state.Key(src))
	if err != nil {
		return false
	}
	val := val0.(*state.VFloat).ToFloat64()
	if val < value {
		return false
	}
	ba := state.MakeVFloat(val - value)
	pool.PutHM("iost", state.Key(src), ba)
	val1, err := pool.GetHM("iost", state.Key(des))
	if val1 == state.VNil {
		ba = state.MakeVFloat(value)
		pool.PutHM("iost", state.Key(des), ba)
	} else {
		val = val1.(*state.VFloat).ToFloat64()
		ba = state.MakeVFloat(val + value)
		pool.PutHM("iost", state.Key(des), ba)
	}
	return true
}
