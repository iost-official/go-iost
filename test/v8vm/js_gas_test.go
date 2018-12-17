package v8vm

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInjectGas(t *testing.T) {
	host, code := MyInit(t, "gas1")
	vmPool.LoadAndCall(host, code, "constructor")

	Convey("test assignment0", t, func() {
		rs, cost, err := vmPool.LoadAndCall(host, code, "assignment0")
		//So(err.Error(), ShouldContainSubstring, "is not iterable")
		So(err, ShouldBeNil)
		t.Log(rs, cost)
	})

	Convey("test assignment1", t, func() {
		rs, cost, err := vmPool.LoadAndCall(host, code, "assignment1")
		So(err.Error(), ShouldContainSubstring, "Cannot assign to read only property")
		t.Log(rs, cost)
	})
	Convey("test assignment11", t, func() {
		rs, cost, err := vmPool.LoadAndCall(host, code, "assignment11")
		So(err, ShouldBeNil)
		t.Log(rs, cost)
	})
	Convey("test assignment111", t, func() {
		rs, cost, err := vmPool.LoadAndCall(host, code, "assignment111")
		So(err, ShouldBeNil)
		t.Log(rs, cost)
	})

	Convey("test assignment2", t, func() {
		rs, cost, err := vmPool.LoadAndCall(host, code, "assignment2", 10)
		So(err, ShouldBeNil)
		t.Log(rs, cost)
		rs, cost, err = vmPool.LoadAndCall(host, code, "assignment2", 1000000)
		So(err, ShouldBeNil)
		t.Log(rs, cost)
	})

	Convey("test assignment3", t, func() {
		rs, cost, err := vmPool.LoadAndCall(host, code, "assignment3", 10)
		So(err.Error(), ShouldContainSubstring, "Arrayconcat is not defined")
		t.Log(rs, cost)
	})

	// deconstruct assignment is not allowed now
	/*
		Convey("test assignment4", t, func(){
			rs, cost, err := vmPool.LoadAndCall(host, code, "assignment4", 10)
			So(err, ShouldBeNil)
			t.Log(rs, cost)
			rs, cost, err = vmPool.LoadAndCall(host, code, "assignment4", 10000)
			So(err, ShouldBeNil)
			t.Log(rs, cost)
		})

		Convey("test assignment44", t, func(){
			rs, cost, err := vmPool.LoadAndCall(host, code, "assignment44", 10)
			So(err, ShouldBeNil)
			t.Log(rs, cost)
		})
		Convey("test assignment444", t, func(){
			rs, cost, err := vmPool.LoadAndCall(host, code, "assignment444", 10)
			So(err, ShouldBeNil)
			t.Log(rs, cost)
		})
		Convey("test assignment4444", t, func(){
			rs, cost, err := vmPool.LoadAndCall(host, code, "assignment4444", 10)
			So(err, ShouldBeNil)
			t.Log(rs, cost)
		})
	*/

	Convey("test instruction counter0", t, func() {
		rs, cost, err := vmPool.LoadAndCall(host, code, "counter0", 10)
		So(err, ShouldBeNil)
		t.Log(rs, cost)
	})

	Convey("test yield0", t, func() {
		rs, cost0, err := vmPool.LoadAndCall(host, code, "yield0", 10)
		So(err, ShouldBeNil)
		So(rs[0], ShouldEqual, "10")
		rs, cost1, err := vmPool.LoadAndCall(host, code, "yield0", 100)
		So(err, ShouldBeNil)
		So(rs[0], ShouldEqual, "100")
		So(cost1.CPU, ShouldBeGreaterThan, cost0.CPU)
	})

	Convey("test yield1", t, func() {
		rs, cost0, err := vmPool.LoadAndCall(host, code, "yield1", 10)
		//So(err.Error(), ShouldContainSubstring, "SyntaxError: Yield expression not allowed in formal parameter")
		So(err, ShouldBeNil)
		t.Log(rs, cost0)
	})

	Convey("test library0", t, func() {
		rs, cost0, err := vmPool.LoadAndCall(host, code, "library0", 10)
		So(err.Error(), ShouldContainSubstring, "esprima is not defined")
		t.Log(rs, cost0)
	})

	Convey("test eval0", t, func() {
		rs, cost0, err := vmPool.LoadAndCall(host, code, "eval0", 10)
		So(err.Error(), ShouldContainSubstring, "Code generation from strings disallowed for this context")
		t.Log(rs, cost0)
	})

	Convey("test function0", t, func() {
		rs, cost0, err := vmPool.LoadAndCall(host, code, "function0", 10)
		So(err.Error(), ShouldContainSubstring, "Function is not a constructor")
		t.Log(rs, cost0)
	})

	Convey("test function1", t, func() {
		rs, cost0, err := vmPool.LoadAndCall(host, code, "function1", 10)
		So(err.Error(), ShouldContainSubstring, "Code generation from strings disallowed for this context")
		t.Log(rs, cost0)
	})

	Convey("test library1", t, func() {
		rs, cost0, err := vmPool.LoadAndCall(host, code, "library1", 10)
		So(err.Error(), ShouldContainSubstring, "a is not a function")
		t.Log(rs, cost0)
	})

	Convey("test literal0", t, func() {
		rs, cost0, err := vmPool.LoadAndCall(host, code, "literal0", 10)
		So(err, ShouldBeNil)
		So(rs[0], ShouldEqual, "token ,nested deeply 10 blah")
		rs, cost1, err := vmPool.LoadAndCall(host, code, "literal0", 100)
		So(err, ShouldBeNil)
		So(rs[0], ShouldEqual, "token ,nested deeply 100 blah")
		So(cost1.CPU, ShouldBeGreaterThan, cost0.CPU)
	})

	Convey("test literal1", t, func() {
		rs, cost0, err := vmPool.LoadAndCall(host, code, "literal1", 10)
		So(err, ShouldBeNil)
		t.Log(rs, cost0)
	})

	Convey("test template string", t, func() {
		rs, cost0, err := vmPool.LoadAndCall(host, code, "templateString", 1, "input")
		So(err, ShouldBeNil)
		t.Log(rs, cost0)
		rs, cost0, err = vmPool.LoadAndCall(host, code, "templateString", 10, "input")
		So(err, ShouldBeNil)
		t.Log(rs, cost0)
	})

	Convey("test length0", t, func() {
		rs, cost0, err := vmPool.LoadAndCall(host, code, "length0", "input")
		So(err, ShouldBeNil)
		t.Log(rs, cost0)
	})

	Convey("test array0", t, func() {
		rs, cost0, err := vmPool.LoadAndCall(host, code, "array0", 1)
		So(err, ShouldBeNil)
		t.Log(rs, cost0)
		rs, cost1, err := vmPool.LoadAndCall(host, code, "array0", 3)
		So(err, ShouldBeNil)
		t.Log(rs, cost0)
		So(cost1.ToGas(), ShouldBeGreaterThan, cost0.ToGas())
	})

	Convey("test string0", t, func() {
		rs, cost0, err := vmPool.LoadAndCall(host, code, "string0", 1)
		So(err, ShouldBeNil)
		t.Log(rs, cost0)
		rs, cost1, err := vmPool.LoadAndCall(host, code, "string0", 3)
		So(err, ShouldBeNil)
		t.Log(rs, cost1)
		So(cost1.ToGas(), ShouldBeGreaterThan, cost0.ToGas())
	})

	Convey("test string1", t, func() {
		rs, cost0, err := vmPool.LoadAndCall(host, code, "string1", 1)
		So(err, ShouldBeNil)
		t.Log(rs, cost0)
		rs, cost1, err := vmPool.LoadAndCall(host, code, "string1", 3)
		So(err, ShouldBeNil)
		t.Log(rs, cost1)
		So(cost1.ToGas(), ShouldBeGreaterThan, cost0.ToGas())
	})

	Convey("test spread0", t, func() {
		rs, cost0, err := vmPool.LoadAndCall(host, code, "spread0", 10)
		So(err, ShouldBeNil)
		t.Log(rs, cost0)
		rs, cost1, err := vmPool.LoadAndCall(host, code, "spread0", 100)
		So(err, ShouldBeNil)
		t.Log(rs, cost1)
		So(cost1.ToGas(), ShouldBeGreaterThan, cost0.ToGas())
	})

	Convey("test bignumber0", t, func() {
		_, cost0, err := vmPool.LoadAndCall(host, code, "bignumber0", "")
		So(err, ShouldBeNil)
		So(cost0.ToGas(), ShouldEqual, int64(241))
	})
}
