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
	//fmt.Print("1 ")
	//fmt.Println(pool.GetHM("iost", state.Key(des)))

	val0, err := pool.GetHM("iost", state.Key(src))
	if err != nil {
		return false
	}
	val := val0.(*state.VFloat).ToFloat64()
	if val < value {
		return false
	}
	ba := state.MakeVFloat(val - value)

	//fmt.Print("1.5 ")
	//fmt.Println(pool.GetHM("iost", state.Key(des)))

	//pool.PutHM("iost", "ahaha", state.MakeVFloat(250))

	pool.PutHM("iost", state.Key(src), ba)

	//fmt.Print("2 ")
	//fmt.Println(pool.GetHM("iost", state.Key(src)))
	//
	//fmt.Print("2.1 ")
	//fmt.Println(pool.GetHM("iost", state.Key(des)))

	val1, err := pool.GetHM("iost", state.Key(des))
	if val1 == state.VNil {
		//fmt.Println("hello")
		ba = state.MakeVFloat(value)
		pool.PutHM("iost", state.Key(des), ba)
	} else {
		val = val1.(*state.VFloat).ToFloat64()
		ba = state.MakeVFloat(val + value)
		pool.PutHM("iost", state.Key(des), ba)
	}
	//fmt.Print("3 ")
	//fmt.Println(pool.GetHM("iost", state.Key(des)))

	return true
}
