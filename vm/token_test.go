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
			_, cost, err := e.LoadAndCall(host, code, "issue", "iost", "user0", "100")
			convey.So(true, convey.ShouldEqual, err.Error() == "token not exists")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(100, 0, 18).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "100")
			convey.So(true, convey.ShouldEqual, err.Error() == "token not exists")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(100, 0, 18).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "issuer0")
			convey.So(true, convey.ShouldEqual, err.Error() == "token not exists")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(100, 0, 18).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "supply", "iost")
			convey.So(true, convey.ShouldEqual, err.Error() == "token not exists")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(100, 0, 18).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "totalSupply", "iost")
			convey.So(true, convey.ShouldEqual, err.Error() == "token not exists")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(100, 0, 18).ToGas())
		})

		convey.Convey("create token without auth", func() {
			_, cost, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err.Error() == "transaction has no permission")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(0, 0, 11).ToGas())
		})

		convey.Convey("create token", func() {
			authList[issuer0] = 1
			host.Context().Set("auth_list", authList)
			_, cost, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(700, 0, 91).ToGas())
		})

		convey.Convey("create duplicate token", func() {
			authList[issuer0] = 1
			host.Context().Set("auth_list", authList)
			_, cost, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(700, 0, 91).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err.Error() == "token exists")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(100, 0, 19).ToGas())
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
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(700, 0, 91).ToGas())
		})

		convey.Convey("issue prepare", func() {
			authList[issuer0] = 1
			host.Context().Set("auth_list", authList)
			_, cost, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(700, 0, 91).ToGas())
		})

		convey.Convey("correct issue", func() {
			_, cost, err := e.LoadAndCall(host, code, "issue", "iost", "user0", "1.1")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(800, 0, 83).ToGas())

			rs, cost, err := e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "1.10000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(400, 0, 42).ToGas())

			rs, cost, err = e.LoadAndCall(host, code, "supply", "iost")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "1.10000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(300, 0, 34).ToGas())

			rs, cost, err = e.LoadAndCall(host, code, "totalSupply", "iost")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "100.00000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(300, 0, 34).ToGas())
		})

		convey.Convey("issue token without auth", func() {
			authList[issuer0] = 0
			host.Context().Set("auth_list", authList)
			_, cost, err := e.LoadAndCall(host, code, "issue", "iost", "user0", "1.1")
			convey.So(true, convey.ShouldEqual, err.Error() == "transaction has no permission")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(400, 0, 43).ToGas())
		})

		convey.Convey("issue too much", func() {
			_, cost, err := e.LoadAndCall(host, code, "issue", "iost", "user0", "1.1")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(800, 0, 83).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "user0", "100")
			convey.So(true, convey.ShouldEqual, err.Error() == "supply too much")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(500, 0, 51).ToGas())

			rs, cost, err := e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "1.10000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(400, 0, 42).ToGas())

			rs, cost, err = e.LoadAndCall(host, code, "supply", "iost")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "1.10000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(300, 0, 34).ToGas())
		})

		convey.Convey("issue invalid amount", func() {
			_, cost, err := e.LoadAndCall(host, code, "issue", "iost", "issuer0", "-1.1")
			convey.So(true, convey.ShouldEqual, err.Error() == "invalid amount")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(500, 0, 51).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "+1.1")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(800, 0, 83).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(500, 0, 51).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "1abc")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(500, 0, 51).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "11111111111111111111111111111111")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(500, 0, 51).ToGas())
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
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(700, 0, 91).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "100.0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(800, 0, 83).ToGas())
		})

		convey.Convey("transfer prepare", func() {
			authList[issuer0] = 1
			host.Context().Set("auth_list", authList)
			_, cost, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(700, 0, 91).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "100.0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(800, 0, 83).ToGas())
		})

		convey.Convey("correct transfer", func() {
			_, cost, err := e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "22.3")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(700, 0, 75).ToGas())

			rs, cost, err := e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "22.30000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(400, 0, 42).ToGas())

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "issuer0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "77.70000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(400, 0, 42).ToGas())

			authList["user0"] = 1
			_, cost, err = e.LoadAndCall(host, code, "transfer", "iost", "user0", "user1", "22.3")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(700, 0, 75).ToGas())

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "0.00000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(400, 0, 42).ToGas())

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user1")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "22.30000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(400, 0, 42).ToGas())

			// transfer to self
			authList["user1"] = 1
			_, cost, err = e.LoadAndCall(host, code, "transfer", "iost", "user1", "user1", "22.3")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(0, 0, 10).ToGas())

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user1")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "22.30000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(400, 0, 42).ToGas())
		})

		convey.Convey("transfer token without auth", func() {
			authList[issuer0] = 0
			host.Context().Set("auth_list", authList)
			_, cost, err := e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "1.1")
			convey.So(true, convey.ShouldEqual, err.Error() == "transaction has no permission")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(100, 0, 19).ToGas())
		})

		convey.Convey("transfer too much", func() {
			_, cost, err := e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "100.1")
			convey.So(true, convey.ShouldEqual, err.Error() == "balance not enough")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(500, 0, 51).ToGas())

			rs, cost, err := e.LoadAndCall(host, code, "balanceOf", "iost", "issuer0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "100.00000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(400, 0, 42).ToGas())

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "0.00000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(300, 0, 34).ToGas())

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user1")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "0.00000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(300, 0, 34).ToGas())
		})

		convey.Convey("transfer invalid amount", func() {
			_, cost, err := e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "-1.1")
			convey.So(true, convey.ShouldEqual, err.Error() == "invalid amount")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(200, 0, 27).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "+1.1")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(700, 0, 75).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(200, 0, 27).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "1abc")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(200, 0, 27).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "transfer", "iost", "issuer0", "user0", "11111111111111111111111111111111")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(200, 0, 27).ToGas())
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
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(700, 0, 91).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "100.0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(800, 0, 83).ToGas())
		})

		convey.Convey("destroy prepare", func() {
			authList[issuer0] = 1
			host.Context().Set("auth_list", authList)
			_, cost, err := e.LoadAndCall(host, code, "create", "iost", "issuer0", int64(100), "{}")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(700, 0, 91).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "issuer0", "100.0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(800, 0, 83).ToGas())
		})

		convey.Convey("correct destroy", func() {
			rs, cost, err := e.LoadAndCall(host, code, "balanceOf", "iost", "issuer0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "100.00000000")

			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(400, 0, 42).ToGas())
			_, cost, err = e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "22.3")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(700, 0, 75).ToGas())

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "issuer0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "77.70000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(400, 0, 42).ToGas())

			rs, cost, err = e.LoadAndCall(host, code, "supply", "iost")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "77.70000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(300, 0, 34).ToGas())

			rs, cost, err = e.LoadAndCall(host, code, "totalSupply", "iost")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "100.00000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(300, 0, 34).ToGas())

		})

		convey.Convey("issuer after destroy", func() {
			_, cost, err := e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "22.3")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(700, 0, 75).ToGas())

			rs, cost, err := e.LoadAndCall(host, code, "balanceOf", "iost", "issuer0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "77.70000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(400, 0, 42).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "user0", "11")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(800, 0, 83).ToGas())

			rs, cost, err = e.LoadAndCall(host, code, "balanceOf", "iost", "user0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "11.00000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(400, 0, 42).ToGas())

			rs, cost, err = e.LoadAndCall(host, code, "supply", "iost")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "88.70000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(300, 0, 34).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "issue", "iost", "user0", "21")
			convey.So(true, convey.ShouldEqual, err.Error() == "supply too much")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(500, 0, 51).ToGas())
		})

		convey.Convey("destroy token without auth", func() {
			authList[issuer0] = 0
			host.Context().Set("auth_list", authList)
			_, cost, err := e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "1.1")
			convey.So(true, convey.ShouldEqual, err.Error() == "transaction has no permission")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(100, 0, 19).ToGas())
		})

		convey.Convey("destroy too much", func() {
			_, cost, err := e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "100.1")
			convey.So(true, convey.ShouldEqual, err.Error() == "balance not enough")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(400, 0, 43).ToGas())

			rs, cost, err := e.LoadAndCall(host, code, "balanceOf", "iost", "issuer0")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "100.00000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(400, 0, 42).ToGas())

			rs, cost, err = e.LoadAndCall(host, code, "supply", "iost")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, len(rs) > 0 && rs[0] == "100.00000000")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(300, 0, 34).ToGas())
		})

		convey.Convey("destroy invalid amount", func() {
			_, cost, err := e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "-1.1")
			convey.So(true, convey.ShouldEqual, err.Error() == "invalid amount")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(200, 0, 27).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "+1.1")
			convey.So(true, convey.ShouldEqual, err == nil)
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(700, 0, 75).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(200, 0, 27).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "1abc")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(200, 0, 27).ToGas())

			_, cost, err = e.LoadAndCall(host, code, "destroy", "iost", "issuer0", "11111111111111111111111111111111")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid")
			convey.So(true, convey.ShouldEqual, cost.ToGas() == contract.NewCost(200, 0, 27).ToGas())
		})
	})
}
