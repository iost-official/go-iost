package v8vm

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInjectGas(t *testing.T) {
	host, code := MyInit(t, "gas1")
	vmPool.LoadAndCall(host, code, "constructor")

	/*
	Convey("test assignment0", t, func(){
		rs, cost, err := vmPool.LoadAndCall(host, code, "assignment0")
		So(err.Error(), ShouldContainSubstring, "is not iterable")
		t.Log(rs, cost)
	})

	Convey("test assignment1", t, func(){
		rs, cost, err := vmPool.LoadAndCall(host, code, "assignment1")
		So(err, ShouldBeNil)
		t.Log(rs, cost)
	})

	Convey("test assignment2", t, func(){
		rs, cost, err := vmPool.LoadAndCall(host, code, "assignment2", 10)
		So(err, ShouldBeNil)
		t.Log(rs, cost)
		rs, cost, err = vmPool.LoadAndCall(host, code, "assignment2", 1000000)
		So(err, ShouldBeNil)
		t.Log(rs, cost)
	})

	Convey("test assignment3", t, func(){
		rs, cost, err := vmPool.LoadAndCall(host, code, "assignment3", 10)
		So(err, ShouldBeNil)
		t.Log(rs, cost)
		rs, cost, err = vmPool.LoadAndCall(host, code, "assignment3", 1000)
		So(err, ShouldBeNil)
		t.Log(rs, cost)
	})
	*/

	Convey("test instruction counter0", t, func(){
		rs, cost, err := vmPool.LoadAndCall(host, code, "counter0", 10)
		So(err, ShouldBeNil)
		t.Log(rs, cost)
		rs, cost, err = vmPool.LoadAndCall(host, code, "counter0", 1000)
		So(err, ShouldBeNil)
		t.Log(rs, cost)
	})
}
