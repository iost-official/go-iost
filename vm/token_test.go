package vm

import (
	"testing"
	"github.com/smartystreets/goconvey/convey"
	"github.com/iost-official/go-iost/core/contract"
	"time"
	"github.com/iost-official/go-iost/vm/native"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
)

func InitVM(t *testing.T, conName string, optional ...interface{}) (*native.Impl, *host.Host, *contract.Contract) {
	db := database.NewDatabaseFromPath(testDataPath + conName + ".json")
	vi := database.NewVisitor(100, db)
	vi.MPut("iost.auth-account", "issuer0", database.MustMarshal(`{"id":"issuer0","permissions":{"active":{"name":"active","groups":[],"users":[{"id":"issuer0","is_key_pair":true,"weight":1}],"threshold":1},"owner":{"name":"owner","groups":[],"users":[{"id":"issuer0","is_key_pair":true,"weight":1}],"threshold":1}}}`))
	vi.MPut("iost.auth-account", "user0", database.MustMarshal(`{"id":"user0","permissions":{"active":{"name":"active","groups":[],"users":[{"id":"user0","is_key_pair":true,"weight":1}],"threshold":1},"owner":{"name":"owner","groups":[],"users":[{"id":"user0","is_key_pair":true,"weight":1}],"threshold":1}}}`))
	vi.MPut("iost.auth-account", "user1", database.MustMarshal(`{"id":"user1","permissions":{"active":{"name":"active","groups":[],"users":[{"id":"user1","is_key_pair":true,"weight":1}],"threshold":1},"owner":{"name":"owner","groups":[],"users":[{"id":"user1","is_key_pair":true,"weight":1}],"threshold":1}}}`))

	ctx := host.NewContext(nil)
	ctx.Set("gas_price", int64(1))
	var gasLimit = int64(10000)
	if len(optional) > 0 {
		gasLimit = optional[0].(int64)
	}
	ctx.GSet("gas_limit", gasLimit)
	ctx.Set("contract_name", conName)
	ctx.Set("tx_hash", []byte("iamhash"))
	ctx.Set("auth_list", make(map[string]int))
	ctx.Set("time", int64(0))

	// pm := NewMonitor()
	h := host.NewHost(ctx, vi, nil, nil)
	h.Context().Set("stack_height", 0)

	code := &contract.Contract{
		ID: "iost.system",
	}

	e := &native.Impl{}
	e.Init()

	return e, h, code
}

func TestToken_Create(t *testing.T) {
	issuer0 := "issuer0"
	e, host, code := InitVM(t, "token")
	code.ID = "iost.token"
	host.SetDeadline(time.Now().Add(10 * time.Second))
	authList := host.Context().Value("auth_list").(map[string]int)

	convey.Convey("Test of Token create", t, func() {
		convey.Reset(func() {
			e, host, code = InitVM(t, "token")
			code.ID = "iost.token"
			host.SetDeadline(time.Now().Add(10 * time.Second))
			authList = host.Context().Value("auth_list").(map[string]int)
		})

		convey.Convey("token not exists", func() {
			authList[issuer0] = 1
			host.Context().Set("auth_list", authList)
			_, _, err := e.LoadAndCall(host, code, "issue", "iost", "user0", "100")
			convey.So(true, convey.ShouldEqual, err.Error() == "token not exists")

			_, _, err = e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "100")
			convey.So(true, convey.ShouldEqual, err.Error() == "token not exists")

			_, _, err = e.LoadAndCall(host, code, "balanceOf", "iost", "issuer0")
			convey.So(true, convey.ShouldEqual, err.Error() == "token not exists")

			_, _, err = e.LoadAndCall(host, code, "supply", "iost")
			convey.So(true, convey.ShouldEqual, err.Error() == "token not exists")

			_, _, err = e.LoadAndCall(host, code, "totalSupply", "iost")
			convey.So(true, convey.ShouldEqual, err.Error() == "token not exists")
		})

		convey.Convey("create token without auth", func() {
			_, _, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err.Error() == "transaction has no permission")
		})

		convey.Convey("create token", func() {
			authList[issuer0] = 1
			host.Context().Set("auth_list", authList)
			_, _, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err == nil)
		})

		convey.Convey("create duplicate token", func() {
			authList[issuer0] = 1
			host.Context().Set("auth_list", authList)
			_, _, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err == nil)

			_, _, err = e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err.Error() == "token exists")
		})
	})
}

func TestToken_Issue(t *testing.T) {
	issuer0 := "issuer0"
	e, host, code := InitVM(t, "token")
	code.ID = "iost.token"
	host.SetDeadline(time.Now().Add(10 * time.Second))
	authList := host.Context().Value("auth_list").(map[string]int)

	convey.Convey("Test of Token issue", t, func() {

		convey.Reset(func() {
			e, host, code = InitVM(t, "token")
			code.ID = "iost.token"
			host.SetDeadline(time.Now().Add(10 * time.Second))
			authList = host.Context().Value("auth_list").(map[string]int)

			authList[issuer0] = 1
			host.Context().Set("auth_list", authList)
			_, cost, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)
		})

		convey.Convey("issue prepare", func() {
			authList[issuer0] = 1
			host.Context().Set("auth_list", authList)
			_, _, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err == nil)
		})

		convey.Convey("correct issue", func() {
			_, cost, err := e.LoadAndCall(host, code, "issue", "iost", "user0", "1.1")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)

			rs, cost, err := e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "1.1")

			rs, cost, err = e.LoadAndCall(host, code, "supply", "iost")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "1.1")

			rs, cost, err = e.LoadAndCall(host, code, "totalSupply", "iost")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "100")
		})

		convey.Convey("issue token without auth", func() {
			delete(authList, issuer0)
			host.Context().Set("auth_list", authList)
			_, _, err := e.LoadAndCall(host, code, "issue", "iost", "user0", "1.1")
			convey.So(true, convey.ShouldEqual, err.Error() == "transaction has no permission")
		})

		convey.Convey("issue too much", func() {
			_, cost, err := e.LoadAndCall(host, code, "issue", "iost", "user0", "1.1")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "user0", "100")
			convey.So(true, convey.ShouldEqual, err.Error() == "supply too much")

			rs, cost, err := e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "1.1")

			rs, cost, err = e.LoadAndCall(host, code, "supply", "iost")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "1.1")
		})

		convey.Convey("issue invalid amount", func() {
			_, cost, err := e.LoadAndCall(host, code, "issue", "iost", "issuer0", "-1.1")
			convey.So(true, convey.ShouldEqual, err.Error() == "invalid amount")
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)

			_, _, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "1.1")
			convey.So(true, convey.ShouldEqual, err == nil)

			_, _, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")

			_, _, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "1abc")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")

			_, _, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "11111111111111111111111111111111")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")
		})

	})
}

func TestToken_Transfer(t *testing.T) {
	issuer0 := "issuer0"
	e, host, code := InitVM(t, "token")
	code.ID = "iost.token"
	host.SetDeadline(time.Now().Add(10 * time.Second))
	authList := host.Context().Value("auth_list").(map[string]int)

	convey.Convey("Test of Token transfer", t, func() {

		convey.Reset(func() {
			e, host, code = InitVM(t, "token")
			code.ID = "iost.token"
			host.SetDeadline(time.Now().Add(10 * time.Second))
			authList = host.Context().Value("auth_list").(map[string]int)

			authList[issuer0] = 1
			host.Context().Set("auth_list", authList)
			_, cost, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "100.0")
			convey.So(true, convey.ShouldEqual, err == nil)
		})

		convey.Convey("transfer prepare", func() {
			authList[issuer0] = 1
			host.Context().Set("auth_list", authList)
			_, cost, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)
			convey.So(true, convey.ShouldEqual, err == nil)

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "100.0")
			convey.So(true, convey.ShouldEqual, err == nil)
		})

		convey.Convey("correct transfer", func() {
			_, cost, err := e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "22.3")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)

			rs, cost, err := e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "22.3")

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "issuer0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "77.7")

			authList["user0"] = 1
			_, cost, err = e.LoadAndCall(host, code, "transfer", "iost", "user0", "user1", "22.3")
			convey.So(true, convey.ShouldEqual, err == nil)

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "0")

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user1")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "22.3")

			// transfer to self
			authList["user1"] = 1
			_, cost, err = e.LoadAndCall(host, code, "transfer", "iost", "user1", "user1", "22.3")
			convey.So(true, convey.ShouldEqual, err == nil)

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user1")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "22.3")
		})

		convey.Convey("transfer token without auth", func() {
			delete(authList, issuer0)
			host.Context().Set("auth_list", authList)
			_, cost, err := e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "1.1")
			convey.So(true, convey.ShouldEqual, err.Error() == "transaction has no permission")
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)
		})

		convey.Convey("transfer too much", func() {
			_, cost, err := e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "100.1")
			convey.So(true, convey.ShouldEqual, err.Error() == "balance not enough")
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)

			rs, cost, err := e.LoadAndCall(host, code, "balanceOf", "iost", "issuer0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "100")

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "0")

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user1")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "0")
		})

		convey.Convey("transfer invalid amount", func() {
			_, cost, err := e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "-1.1")
			convey.So(true, convey.ShouldEqual, err.Error() == "invalid amount")
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)

			_, cost, err = e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "1.1")
			convey.So(true, convey.ShouldEqual, err == nil)

			_, cost, err = e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")

			_, cost, err = e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "1abc")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")

			_, cost, err = e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "11111111111111111111111111111111")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")
		})
	})
}

func TestToken_Destroy(t *testing.T) {
	issuer0 := "issuer0"
	e, host, code := InitVM(t, "token")
	code.ID = "iost.token"
	host.SetDeadline(time.Now().Add(10 * time.Second))
	authList := host.Context().Value("auth_list").(map[string]int)

	convey.Convey("Test of Token destroy", t, func() {

		convey.Reset(func() {
			e, host, code = InitVM(t, "token")
			code.ID = "iost.token"
			host.SetDeadline(time.Now().Add(10 * time.Second))
			authList = host.Context().Value("auth_list").(map[string]int)

			authList[issuer0] = 1
			host.Context().Set("auth_list", authList)
			_, cost, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "100.0")
			convey.So(true, convey.ShouldEqual, err == nil)
		})

		convey.Convey("destroy prepare", func() {
			authList[issuer0] = 1
			host.Context().Set("auth_list", authList)
			_, cost, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "100.0")
			convey.So(true, convey.ShouldEqual, err == nil)
		})

		convey.Convey("correct destroy", func() {
			rs, cost, err := e.LoadAndCall(host, code, "balanceOf", "iost", "issuer0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "100")
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)

			_, cost, err = e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "22.3")
			convey.So(true, convey.ShouldEqual, err == nil)

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "issuer0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "77.7")

			rs, cost, err = e.LoadAndCall(host, code, "supply", "iost")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "77.7")

			rs, cost, err = e.LoadAndCall(host, code, "totalSupply", "iost")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "100")

		})

		convey.Convey("issuer after destroy", func() {
			_, cost, err := e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "22.3")
			convey.So(true, convey.ShouldEqual, err == nil)

			rs, cost, err := e.LoadAndCall(host, code, "balanceOf", "iost", "issuer0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "77.7")

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "user0", "11")
			convey.So(true, convey.ShouldEqual, err == nil)

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "11")

			rs, cost, err = e.LoadAndCall(host, code, "supply", "iost")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "88.7")
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "user0", "21")
			convey.So(true, convey.ShouldEqual, err.Error() == "supply too much")
		})

		convey.Convey("destroy token without auth", func() {
			delete(authList, issuer0)
			host.Context().Set("auth_list", authList)
			_, _, err := e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "1.1")
			convey.So(true, convey.ShouldEqual, err.Error() == "transaction has no permission")
		})

		convey.Convey("destroy too much", func() {
			_, cost, err := e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "100.1")
			convey.So(true, convey.ShouldEqual, err.Error() == "balance not enough")

			rs, cost, err := e.LoadAndCall(host, code, "balanceOf", "iost", "issuer0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "100")

			rs, cost, err = e.LoadAndCall(host, code, "supply", "iost")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "100")
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)
		})

		convey.Convey("destroy invalid amount", func() {
			_, cost, err := e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "-1.1")
			convey.So(true, convey.ShouldEqual, err.Error() == "invalid amount")
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)

			_, cost, err = e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "1.1")
			convey.So(true, convey.ShouldEqual, err == nil)

			_, cost, err = e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")

			_, cost, err = e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "1abc")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")

			_, cost, err = e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "11111111111111111111111111111111")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")
		})
	})
}

func TestToken_TransferFreeze(t *testing.T) {
	issuer0 := "issuer0"
	e, host, code := InitVM(t, "token")
	code.ID = "iost.token"
	host.SetDeadline(time.Now().Add(10 * time.Second))
	authList := host.Context().Value("auth_list").(map[string]int)
	now := int64(time.Now().Unix())

	convey.Convey("Test of Token transferFreeze", t, func() {

		convey.Reset(func() {
			e, host, code = InitVM(t, "token")
			code.ID = "iost.token"
			host.SetDeadline(time.Now().Add(10 * time.Second))
			authList = host.Context().Value("auth_list").(map[string]int)

			authList[issuer0] = 1
			host.Context().Set("auth_list", authList)
			_, _, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err == nil)

			_, _, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "100.0")
			convey.So(true, convey.ShouldEqual, err == nil)
		})

		convey.Convey("transferFreeze prepare", func() {
			authList[issuer0] = 1
			host.Context().Set("auth_list", authList)
			_, cost, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "100.0")
			convey.So(true, convey.ShouldEqual, err == nil)
		})

		convey.Convey("correct transferFreeze", func() {
			_, cost, err := e.LoadAndCall(host, code, "transferFreeze", "iost", "issuer0", "user0", "22.3", now)
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)

			rs, cost, err := e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "0")

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "issuer0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "77.7")

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "0")

			host.Context().Set("time", now + 1)

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "22.3")

			// transferFreeze to self
			authList["user0"] = 1
			_, cost, err = e.LoadAndCall(host, code, "transferFreeze", "iost", "user0", "user0", "10", now + 10)
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() > 0)

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "12.3")

			_, cost, err = e.LoadAndCall(host, code, "transferFreeze", "iost", "user0", "user0", "1", now + 20)
			convey.So(true, convey.ShouldEqual, err == nil)

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "11.3")

			host.Context().Set("time", now + 11)
			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "21.3")

			host.Context().Set("time", now + 21)
			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "22.3")
		})

		convey.Convey("transferFreeze token without auth", func() {
			delete(authList, issuer0)
			host.Context().Set("auth_list", authList)
			_, _, err := e.LoadAndCall(host, code, "transferFreeze", "iost", "issuer0", "user0", "1.1", now)
			convey.So(true, convey.ShouldEqual, err.Error() == "transaction has no permission")
		})

		convey.Convey("transferFreeze too much", func() {
			_, _, err := e.LoadAndCall(host, code, "transferFreeze", "iost", "issuer0", "user0", "100.1", now - 1)
			convey.So(true, convey.ShouldEqual, err.Error() == "balance not enough")

			rs, _, err := e.LoadAndCall(host, code, "balanceOf", "iost", "issuer0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "100")

			rs, _, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "0")

			_, _, err = e.LoadAndCall(host, code, "transferFreeze", "iost", "issuer0", "user0", "100", now + 100)
			convey.So(true, convey.ShouldEqual, err == nil)

			authList["user0"] = 1
			host.Context().Set("auth_list", authList)
			_, _, err = e.LoadAndCall(host, code, "transferFreeze", "iost", "user0", "user1", "10", now + 100)
			convey.So(true, convey.ShouldEqual, err.Error() == "balance not enough")

			_, _, err = e.LoadAndCall(host, code, "transfer", "iost", "user0", "user1", "10")
			convey.So(true, convey.ShouldEqual, err.Error() == "balance not enough")
		})

	})
}
