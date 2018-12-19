package integration

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/event"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	. "github.com/iost-official/go-iost/verifier"
)

func Test_callWithAuth(t *testing.T) {
	ilog.Stop()
	Convey("test of callWithAuth", t, func() {
		s := NewSimulator()
		defer s.Clear()

		acc := testAccounts[0]
		createAccountsWithResource(s)
		createToken(t, s, acc)

		ca, err := s.Compile("Contracttransfer", "./test_data/transfer", "./test_data/transfer.js")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		cname, r, err := s.DeployContract(ca, acc.ID, acc.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		Convey("test of callWithAuth", func() {
			s.Visitor.SetTokenBalanceFixed("iost", cname, "1000")
			r, err := s.Call(cname, "withdraw", fmt.Sprintf(`["%v", "%v"]`, acc.ID, "10"), acc.ID, acc.KeyPair)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			balance := common.Fixed{Value: s.Visitor.TokenBalance("iost", cname), Decimal: s.Visitor.Decimal("iost")}
			So(balance.ToString(), ShouldEqual, "990")
		})
	})
}

func Test_VMMethod(t *testing.T) {
	ilog.Stop()
	Convey("test of vm method", t, func() {
		s := NewSimulator()
		defer s.Clear()

		acc := testAccounts[0]
		createAccountsWithResource(s)
		createToken(t, s, acc)

		ca, err := s.Compile("", "./test_data/vmmethod", "./test_data/vmmethod")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		cname, r, err := s.DeployContract(ca, acc.ID, acc.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		Convey("test of contract name", func() {
			r, err := s.Call(cname, "contractName", "[]", acc.ID, acc.KeyPair)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(len(r.Returns), ShouldEqual, 1)
			res, err := json.Marshal([]interface{}{cname})
			So(err, ShouldBeNil)
			So(r.Returns[0], ShouldEqual, string(res))
		})

		Convey("test of receipt", func() {
			r, err := s.Call(cname, "receiptf", fmt.Sprintf(`["%v"]`, "receiptdata"), acc.ID, acc.KeyPair)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(len(r.Receipts), ShouldEqual, 1)
			So(r.Receipts[0].Content, ShouldEqual, "receiptdata")
			So(r.Receipts[0].FuncName, ShouldEqual, cname+"/receiptf")
		})

	})
}

func Test_VMMethod_Event(t *testing.T) {
	ilog.Stop()
	Convey("test of vm method event", t, func() {
		s := NewSimulator()
		defer s.Clear()
		
		acc := testAccounts[0]
		createAccountsWithResource(s)

		ca, err := s.Compile("", "./test_data/vmmethod", "./test_data/vmmethod")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		cname, r, err := s.DeployContract(ca, acc.ID, acc.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		eve := event.GetCollector()
		// contract event
		ch1 := eve.Subscribe(1, []event.Topic{event.ContractEvent}, nil)
		ch2 := eve.Subscribe(2, []event.Topic{event.ContractReceipt}, nil)

		r, err = s.Call(cname, "event", fmt.Sprintf(`["%v"]`, "eventdata"), acc.ID, acc.KeyPair)
		s.Visitor.Commit()

		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")

		e := <-ch1
		So(e.Data, ShouldEqual, "eventdata")
		So(e.Topic, ShouldEqual, event.ContractEvent)

		// receipt event
		r, err = s.Call(cname, "receiptf", fmt.Sprintf(`["%v"]`, "receipteventdata"), acc.ID, acc.KeyPair)
		s.Visitor.Commit()

		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")

		e = <-ch2
		So(e.Data, ShouldEqual, "receipteventdata")
		So(e.Topic, ShouldEqual, event.ContractReceipt)
	})
}

func Test_RamPayer(t *testing.T) {
	ilog.Stop()
	Convey("test of ram payer", t, func() {
		s := NewSimulator()
		defer s.Clear()

		acc := testAccounts[0]
		createAccountsWithResource(s)
		createToken(t, s, acc)

		ca, err := s.Compile("", "./test_data/vmmethod", "./test_data/vmmethod")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		cname, r, err := s.DeployContract(ca, acc.ID, acc.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		Convey("test of put and get", func() {
			//ram := s.GetRAM(acc.ID)
			r, err := s.Call(cname, "putwithpayer", fmt.Sprintf(`["k", "v", "%v"]`, acc.ID), acc.ID, acc.KeyPair)
			s.Visitor.Commit()
			So(s.GetRAM(acc.ID), ShouldEqual, 8482)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)

			r, err = s.Call(cname, "get", fmt.Sprintf(`["k"]`), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(len(r.Returns), ShouldEqual, 1)
			So(r.Returns[0], ShouldEqual, "[\"v\"]")
		})

		Convey("test of map put and get", func() {
			//ram := s.GetRAM(acc.ID)
			r, err := s.Call(cname, "mapputwithpayer", fmt.Sprintf(`["k", "f", "v", "%v"]`, acc.ID), acc.ID, acc.KeyPair)
			s.Visitor.Commit()
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.GetRAM(acc.ID), ShouldEqual, 8480)

			r, err = s.Call(cname, "mapget", fmt.Sprintf(`["k", "f"]`), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(len(r.Returns), ShouldEqual, 1)
			So(r.Returns[0], ShouldEqual, "[\"v\"]")
		})

		Convey("test of map put and get change payer", func() {
			acc2 := testAccounts[1]
			//ram := s.GetRAM(acc.ID)
			r, err := s.Call(cname, "mapputwithpayer", fmt.Sprintf(`["k", "f", "vv", "%v"]`, acc.ID), acc.ID, acc.KeyPair)
			s.Visitor.Commit()
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.GetRAM(acc.ID), ShouldEqual, 8479)

			//ram = s.GetRAM(acc.ID)
			ram1 := s.GetRAM(acc2.ID)
			r, err = s.Call(cname, "mapputwithpayer", fmt.Sprintf(`["k", "f", "vvv", "%v"]`, acc2.ID), acc2.ID, acc2.KeyPair)
			s.Visitor.Commit()
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.GetRAM(acc.ID), ShouldEqual, 8545)
			So(s.GetRAM(acc2.ID), ShouldEqual, 9933)

			ram1 = s.GetRAM(acc2.ID)
			r, err = s.Call(cname, "mapputwithpayer", fmt.Sprintf(`["k", "f", "v", "%v"]`, acc2.ID), acc2.ID, acc2.KeyPair)
			s.Visitor.Commit()
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.GetRAM(acc2.ID), ShouldEqual, ram1+2)

			ram1 = s.GetRAM(acc2.ID)
			r, err = s.Call(cname, "mapputwithpayer", fmt.Sprintf(`["k", "f", "vvvvv", "%v"]`, acc2.ID), acc2.ID, acc2.KeyPair)
			s.Visitor.Commit()
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.GetRAM(acc2.ID), ShouldEqual, ram1-4)
		})

		Convey("test nested call check payer", func() {
			ram0 := s.GetRAM(acc.ID)
			acc3 := testAccounts[2]
			acc4 := testAccounts[3]
			if err != nil {
				t.Fatal(err)
			}
			ca, err := s.Compile("", "./test_data/nest0", "./test_data/nest0")
			if err != nil || ca == nil {
				t.Fatal(err)
			}
			cname0, r, err := s.DeployContract(ca, acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)

			ca, err = s.Compile("", "./test_data/nest1", "./test_data/nest1")
			if err != nil || ca == nil {
				t.Fatal(err)
			}
			cname1, r, err := s.DeployContract(ca, acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)

			So(s.GetRAM(acc.ID), ShouldEqual, ram0-2533)

			ram0 = s.GetRAM(acc.ID)
			//ram4 := s.GetRAM(acc3.ID)
			ram6 := s.GetRAM(acc4.ID)
			s.Visitor.SetTokenBalanceFixed("iost", acc3.ID, "100")
			r, err = s.Call(cname0, "call", fmt.Sprintf(`["%v", "test", "%v"]`, cname1,
				fmt.Sprintf(`[\"%v\", \"%v\"]`, acc3.ID, acc4.ID)), acc3.ID, acc3.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)

			So(s.GetRAM(acc4.ID), ShouldEqual, ram6)
			So(s.GetRAM(acc3.ID), ShouldEqual, 9957)
			So(s.GetRAM(acc.ID), ShouldEqual, ram0-6)
		})
	})
}

func Test_StackHeight(t *testing.T) {
	ilog.Stop()
	Convey("test of stack height", t, func() {
		s := NewSimulator()
		defer s.Clear()

		acc := testAccounts[0]
		createAccountsWithResource(s)
		createToken(t, s, acc)

		ca, err := s.Compile("", "./test_data/nest0", "./test_data/nest0")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		cname0, r, err := s.DeployContract(ca, acc.ID, acc.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		ca, err = s.Compile("", "./test_data/nest1", "./test_data/nest1")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		cname1, r, err := s.DeployContract(ca, acc.ID, acc.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		Convey("test of out of stack height", func() {
			r, err := s.Call(cname0, "sh0", fmt.Sprintf(`["%v"]`, cname1), acc.ID, acc.KeyPair)
			s.Visitor.Commit()
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldContainSubstring, "stack height exceed.")
		})
	})
}

func Test_Validate(t *testing.T) {
	ilog.Stop()
	Convey("test validate", t, func() {
		s := NewSimulator()
		defer s.Clear()
		acc := prepareAuth(t, s)
		s.SetAccount(acc.ToAccount())
		s.SetGas(acc.ID, 10000000)
		s.SetRAM(acc.ID, 300)

		c, err := s.Compile("validate", "test_data/validate", "test_data/validate")
		So(err, ShouldBeNil)
		So(len(c.Encode()), ShouldEqual, 133)
		_, r, err := s.DeployContract(c, acc.ID, acc.KeyPair)
		s.Visitor.Commit()
		So(err.Error(), ShouldContainSubstring, "abi not defined in source code: c")
		So(r.Status.Message, ShouldEqual, "validate code error: , result: Error: abi not defined in source code: c")

		c, err = s.Compile("validate1", "test_data/validate1", "test_data/validate1")
		So(err, ShouldBeNil)
		_, r, err = s.DeployContract(c, acc.ID, acc.KeyPair)
		s.Visitor.Commit()
		So(err.Error(), ShouldContainSubstring, "Error: args should be one of ")
		So(r.Status.Message, ShouldContainSubstring, "validate code error: , result: Error: args should be one of ")
	})
}

func Test_SpecialChar(t *testing.T) {
	ilog.Stop()
	spStr := ""
	for i := 0x00; i <= 0x1F; i++ {
		spStr += fmt.Sprintf("const char%d = `%s`;\n", i, string(rune(i)))
	}
	spStr += fmt.Sprintf("const char%d = `%s`;\n", 0x7F, string(rune(0x7F)))
	for i := 0x80; i <= 0x9F; i++ {
		spStr += fmt.Sprintf("const char%d = `%s`;\n", i, string(rune(i)))
	}
	spStr += fmt.Sprintf("const char%d = `%s`;\n", 0x2028, string(rune(0x2028)))
	spStr += fmt.Sprintf("const char%d = `%s`;\n", 0x2029, string(rune(0x2029)))
	spStr += fmt.Sprintf("const char%d = `%s`;\n", 0xE0001, string(rune(0xE0001)))
	for i := 0xE0020; i <= 0xE007F; i++ {
		spStr += fmt.Sprintf("const char%d = `%s`;\n", i, string(rune(i)))
	}
	lst := []int64{0x061C, 0x200E, 0x200F, 0x202A, 0x202B, 0x202C, 0x202D, 0x202E, 0x2066, 0x2067, 0x2068, 0x2069}
	for _, i := range lst {
		spStr += fmt.Sprintf("const char%d = `%s`;\n", i, string(rune(i)))
	}
	for i := 0xE0100; i <= 0xE01EF; i++ {
		spStr += fmt.Sprintf("const char%d = `%s`;\n", i, string(rune(i)))
	}
	for i := 0x180B; i <= 0x180E; i++ {
		spStr += fmt.Sprintf("const char%d = `%s`;\n", i, string(rune(i)))
	}
	for i := 0x200C; i <= 0x200D; i++ {
		spStr += fmt.Sprintf("const char%d = `%s`;\n", i, string(rune(i)))
	}
	for i := 0xFFF0; i <= 0xFFFF; i++ {
		spStr += fmt.Sprintf("const char%d = `%s`;\n", i, string(rune(i)))
	}
	code := spStr +
		"class Test {" +
		"	init() {}" +
		"	transfer(from, to, amountJson) {" +
		"		blockchain.transfer(from, to, amountJson.amount, '');" +
		"	}" +
		"};" +
		"module.exports = Test;"

	abi := `
	{
		"lang": "javascript",
		"version": "1.0.0",
		"abi": [
			{
				"name": "transfer",
				"args": [
					"string",
					"string",
					"json"
				]
			}
		]
	}
	`
	Convey("test validate", t, func() {
		s := NewSimulator()
		defer s.Clear()
		acc := prepareAuth(t, s)
		createAccountsWithResource(s)
		createToken(t, s, acc)
		s.SetGas(acc.ID, 10000000)
		s.SetRAM(acc.ID, 100000)

		c, err := (&contract.Compiler{}).Parse("", code, abi)
		So(err, ShouldBeNil)

		cname, _, err := s.DeployContract(c, acc.ID, acc.KeyPair)
		s.Visitor.Commit()
		So(err, ShouldBeNil)

		acc2 := testAccounts[1]
		s.Visitor.SetTokenBalanceFixed("iost", acc.ID, "1000")
		s.Visitor.SetTokenBalanceFixed("iost", acc2.ID, "1000")
		params := []interface{}{
			acc.ID,
			acc2.ID,
			map[string]string{
				"amount": "1000",
				"hack":   "\u2028\u2029\u0000",
			},
		}
		paramsByte, err := json.Marshal(params)
		So(err, ShouldBeNil)
		r, err := s.Call(cname, "transfer", string(paramsByte), acc.ID, acc.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)
		So(s.Visitor.TokenBalanceFixed("iost", acc2.ID).ToString(), ShouldEqual, "2000")
	})
}

func Test_CallResult(t *testing.T) {
	ilog.Stop()
	Convey("test call result", t, func() {
		s := NewSimulator()
		defer s.Clear()
		acc := prepareAuth(t, s)
		s.SetAccount(acc.ToAccount())
		s.SetGas(acc.ID, 2000000)
		s.SetRAM(acc.ID, 10000)

		c, err := s.Compile("", "test_data/callresult", "test_data/callresult")
		So(err, ShouldBeNil)
		cname, r, err := s.DeployContract(c, acc.ID, acc.KeyPair)
		//s.Visitor.Put(cname+"-ret", database.MustMarshal("ab\x00c"))
		s.Visitor.Commit()
		So(err, ShouldBeNil)
		r, err = s.Call(cname, "ret_eof", `[]`, acc.ID, acc.KeyPair)
		So(err, ShouldBeNil)
		a := s.Visitor.Get(cname + "-ret")
		b := strings.NewReplacer("\x00", "\\x00").Replace(a)
		So(b, ShouldEqual, "sab\\x00c@"+cname)
		So(len(r.Returns), ShouldEqual, 1)
		So(r.Returns[0], ShouldEqual, `["ab\u0000cd"]`)
	})
}

func Test_ReturnObjectToJsonError(t *testing.T) {
	ilog.Stop()
	Convey("test return object toJSON error", t, func() {
		s := NewSimulator()
		defer s.Clear()
		acc := prepareAuth(t, s)
		s.SetAccount(acc.ToAccount())
		s.SetGas(acc.ID, 2000000)
		s.SetRAM(acc.ID, 10000)

		c, err := s.Compile("", "test_data/callresult", "test_data/callresult")
		So(err, ShouldBeNil)
		cname, r, err := s.DeployContract(c, acc.ID, acc.KeyPair)
		s.Visitor.Commit()
		So(err, ShouldBeNil)

		r, err = s.Call(cname, "ret_obj", `[]`, acc.ID, acc.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.ErrorRuntime)
		So(r.Status.Message, ShouldContainSubstring, "error in JSON.stringfy")
	})
}
