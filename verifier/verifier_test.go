package verifier

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/core/mocks"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	"github.com/iost-official/prototype/vm/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

// nolint
func TestGenesisVerify(t *testing.T) {
	Convey("Test of Genesis verify", t, func() {
		Convey("Parse contract", func() {
			mockCtl := gomock.NewController(t)
			pool := core_mock.NewMockPool(mockCtl)
			var count int
			var k, f, k2 state.Key
			var v, v2 state.Value
			pool.EXPECT().PutHM(gomock.Any(), gomock.Any(), gomock.Any()).Times(2).Do(func(key, field state.Key, value state.Value) error {
				k, f, v = key, field, value
				count++
				return nil
			})
			pool.EXPECT().Put(gomock.Any(), gomock.Any()).Do(func(key state.Key, value state.Value) {
				k2, v2 = key, value
			})
			pool.EXPECT().Copy().Return(pool)
			contract := vm_mock.NewMockContract(mockCtl)
			contract.EXPECT().Code().Return(`
-- @PutHM iost abc f10000
-- @PutHM iost def f1000
-- @Put hello sworld
`)
			_, err := ParseGenesis(contract, pool)
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 2)
			So(k, ShouldEqual, state.Key("iost"))
			So(v2.EncodeString(), ShouldEqual, "sworld")

		})
	})
}

// nolint
func TestCacheVerifier(t *testing.T) {
	Convey("Test of CacheVerifier", t, func() {
		Convey("Verify contract", func() {
			mockCtl := gomock.NewController(t)
			pool := core_mock.NewMockPool(mockCtl)

			var k state.Key
			var v state.Value

			pool.EXPECT().Put(gomock.Any(), gomock.Any()).AnyTimes().Do(func(key state.Key, value state.Value) error {
				k = key
				v = value
				return nil
			})

			pool.EXPECT().Get(gomock.Any()).AnyTimes().Return(state.MakeVFloat(3.14), nil)

			var k2 state.Key
			var f2 state.Key
			var v2 state.Value
			pool.EXPECT().PutHM(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(key, field state.Key, value state.Value) {
				k2 = key
				f2 = field
				v2 = value
			})
			v3 := state.MakeVFloat(float64(1000000))
			pool.EXPECT().GetHM(gomock.Any(), gomock.Any()).AnyTimes().Return(v3, nil)
			pool.EXPECT().Copy().AnyTimes().Return(pool)
			main := lua.NewMethod(vm.Public, "main", 0, 1)
			code := `function main()
	a = Get("pi")
	Put("hello", a)
	return "success"
end`
			lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 10000, Price: 1, Publisher: vm.IOSTAccount("ahaha")}, code, main)

			cv := NewCacheVerifier()
			_, err := cv.VerifyContract(&lc, pool)
			So(err, ShouldBeNil)
			So(string(k), ShouldEqual, "testhello")
			So(v.EncodeString(), ShouldEqual, "true")
			So(string(k2), ShouldEqual, "iost")
			So(string(f2), ShouldEqual, "ahaha")
			vv := v2.(*state.VFloat)
			So(vv.ToFloat64(), ShouldEqual, float64(997989.99))
		})
		Convey("Verify free contract", func() {
			mockCtl := gomock.NewController(t)
			pool := core_mock.NewMockPool(mockCtl)

			var k state.Key
			var v state.Value

			pool.EXPECT().Put(gomock.Any(), gomock.Any()).AnyTimes().Do(func(key state.Key, value state.Value) error {
				k = key
				v = value
				return nil
			})

			pool.EXPECT().Get(gomock.Any()).AnyTimes().Return(state.MakeVFloat(3.14), nil)

			var k2 state.Key
			var f2 state.Key
			var v2 state.Value
			pool.EXPECT().PutHM(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(key, field state.Key, value state.Value) {
				k2 = key
				f2 = field
				v2 = value
			})
			//v3 := state.MakeVFloat(float64(10000))
			pool.EXPECT().GetHM(gomock.Any(), gomock.Any()).AnyTimes().Return(state.MakeVFloat(1000000), nil)
			pool.EXPECT().Copy().AnyTimes().Return(pool)
			main := lua.NewMethod(vm.Public, "main", 0, 1)
			code := `function main()
	a = Get("pi")
	Put("hello", a)
	return "success"
end`
			lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 10000, Price: 0, Publisher: vm.IOSTAccount("ahaha")}, code, main)

			cv := NewCacheVerifier()
			_, err := cv.VerifyContract(&lc, pool)
			So(err, ShouldBeNil)
			So(string(k), ShouldEqual, "testhello")
			So(v.EncodeString(), ShouldEqual, "true")
		})
	})
}

func TestCacheVerifier_TransferOnly(t *testing.T) {
	Convey("System test of transfer", t, func() {
		main := lua.NewMethod(vm.Public, "main", 0, 1)
		code := `function main()
	Transfer("a", "b", 50)
end`
		lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 10000, Price: 1, Publisher: vm.IOSTAccount("a")}, code, main)

		dbx, err := db.DatabaseFactory("redis")
		if err != nil {
			panic(err.Error())
		}
		sdb := state.NewDatabase(dbx)
		pool := state.NewPool(sdb)
		pool.PutHM(state.Key("iost"), state.Key("a"), state.MakeVFloat(1000000))
		pool.PutHM(state.Key("iost"), state.Key("b"), state.MakeVFloat(1000000))
		var pool2 state.Pool

		cv := NewCacheVerifier()
		pool2, err = cv.VerifyContract(&lc, pool)
		So(err, ShouldBeNil)
		aa, err := pool2.GetHM("iost", "a")
		ba, err := pool2.GetHM("iost", "b")
		So(err, ShouldBeNil)
		So(aa.(*state.VFloat).ToFloat64(), ShouldEqual, 999943.99)
		So(ba.(*state.VFloat).ToFloat64(), ShouldEqual, 1000050)
	})

}

func TestCacheVerifier_Multiple(t *testing.T) {
	Convey("System test of transfer", t, func() {
		main := lua.NewMethod(vm.Public, "main", 0, 1)
		code := `function main()
	Transfer("a", "b", 50)
end`
		lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 10000, Price: 1, Publisher: vm.IOSTAccount("a")}, code, main)

		dbx, err := db.DatabaseFactory("redis")
		if err != nil {
			panic(err.Error())
		}
		sdb := state.NewDatabase(dbx)
		pool := state.NewPool(sdb)
		pool.PutHM(state.Key("iost"), state.Key("a"), state.MakeVFloat(1000000))
		pool.PutHM(state.Key("iost"), state.Key("b"), state.MakeVFloat(1000000))
		var pool2 state.Pool

		cv := NewCacheVerifier()
		pool2, err = cv.VerifyContract(&lc, pool)
		if err != nil {
			panic(err)
		}
		pool3, err := cv.VerifyContract(&lc, pool2)
		if err != nil {
			panic(err)
		}
		_, err = pool2.GetHM("iost", "a")
		ba, err := pool2.GetHM("iost", "b")

		aa2, err := pool3.GetHM("iost", "a")
		So(err, ShouldBeNil)
		So(aa2.(*state.VFloat).ToFloat64(), ShouldEqual, 999887.98)
		So(ba.(*state.VFloat).ToFloat64(), ShouldEqual, 1000100)
	})

}

func BenchmarkCacheVerifier_TransferOnly(b *testing.B) {
	main := lua.NewMethod(vm.Public, "main", 0, 1)
	code := `function main()
	Transfer("a", "b", 50)
    return "success"
end`
	lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 10000, Price: 1, Publisher: vm.IOSTAccount("a")}, code, main)

	dbx, err := db.DatabaseFactory("redis")
	if err != nil {
		panic(err.Error())
	}
	sdb := state.NewDatabase(dbx)
	pool := state.NewPool(sdb)
	pool.PutHM(state.Key("iost"), state.Key("a"), state.MakeVFloat(1000000))
	pool.PutHM(state.Key("iost"), state.Key("b"), state.MakeVFloat(1000000))

	var pool2 state.Pool

	cv := NewCacheVerifier()
	for i := 0; i < b.N; i++ {
		pool2, err = cv.VerifyContract(&lc, pool)
		if err != nil {
			panic(err)
		}
	}

	_ = pool2

}

func BenchmarkCacheVerifierWithCache_TransferOnly(b *testing.B) {
	main := lua.NewMethod(vm.Public, "main", 0, 1)
	code := `function main()
	Transfer("a", "b", 50)
    return "success"
end`
	lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 10000, Price: 1, Publisher: vm.IOSTAccount("a")}, code, main)

	dbx, err := db.DatabaseFactory("redis")
	if err != nil {
		panic(err.Error())
	}
	sdb := state.NewDatabase(dbx)
	pool := state.NewPool(sdb)
	pool.PutHM(state.Key("iost"), state.Key("a"), state.MakeVFloat(1000000))
	pool.PutHM(state.Key("iost"), state.Key("b"), state.MakeVFloat(1000000))

	var pool2 state.Pool

	cv := NewCacheVerifier()
	for i := 0; i < b.N; i++ {
		pool2, err = cv.VerifyContract(&lc, pool)
		if err != nil {
			panic(err)
		}
		pool = pool2
	}

	_ = pool2

}

func BenchmarkLuckyBet(b *testing.B) {
	buy := `--- main 合约主入口
-- LuckyBet
-- @gas_limit 100000000
-- @gas_price 0
-- @param_cnt 0
-- @return_cnt 0
function main()
	ok, r = Call("main", "Bet", "caller", 6, 4, 14238)
	Log(string.format("bet %s", tostring(ok)))
	Log(string.format("bet r = %s", tostring(r)))
	Assert(ok)
	Assert(r == 0)
end--f`

	main := `--- main 猜区块
-- bet your block and get reward!
-- @gas_limit 100000000
-- @gas_price 0
-- @param_cnt 0
-- @return_cnt 0
-- @publisher walleta
function main()
	Assert(Put("max_user_number", 100))
	Assert(Put("user_number", 0))
	Assert(Put("total_coins", 0))
	Assert(Put("last_lucky_block", -1))
	Assert(Put("round", 0))
	Assert(clearUserValue() == 0)
end--f

--- clearUserValue clear user bet value 
-- @param_cnt 0
-- @return_cnt 1
-- @privilege private
function clearUserValue()
	clearTable = {}
	for i = 0, 9, 1 do
		userTableKey = string.format("user_value%d", i)
		Log(string.format("clear user value : %s", userTableKey))
		print(string.format("clear user value : %s", userTableKey))
		ok, json = ToJson(clearTable)
		Assert(ok)
		Assert(Put(userTableKey, json))
		ok, tmpValue = Get(userTableKey)
		Assert(ok)
		Log(string.format("cleared user value : %v", tmpValue))
		print(string.format("cleared user value : %v", tmpValue))
	end
	return 0
end--f

--- Bet a lucky number
-- bet a lucky number with 1 ~ 5 coins
-- @param_cnt 4
-- @return_cnt 1
-- @privilege public
function Bet(account, luckyNumber, coins, nonce)
	if (not (coins >= 1 and coins <= 5))
	then
	    return "bet coins should be >=1 and <= 5"
	end
	if (not (luckyNumber >= 0 and luckyNumber <= 9))
	then
	    return "bet lucky number should be >=0 and <= 9"
	end
	Log(string.format("before account = %s, lucky = %d, coin = %f, nonce = %d", account, luckyNumber, coins, nonce))
	Assert(nonce ~= nil)

	ok, maxUserNumber = Get("max_user_number")
	Assert(ok)
    ok, number = Get("user_number")
	Assert(ok)
    ok, totalCoins = Get("total_coins")
	Assert(ok)

	Log(string.format("account = %s, lucky = %d, coin = %f, nonce = %d", account, luckyNumber, coins, nonce))

	Assert(Deposit(account, coins) == true)
	userTableKey = string.format("user_value%d", luckyNumber)
	Log("after deposit, usertablekey = "..userTableKey)

	ok, json = Get(userTableKey)
	Assert(ok)
	Log(string.format("val json: %v", json))
	ok, valTable = ParseJson(json)
	Assert(ok)
	Log(string.format("val table: %v", valTable))
	if (valTable == nil)
	then
		valTable = {}
	end
	Assert(valTable ~= nil)
	Log("val table not nil")

	len = #valTable
	valTable[len + 1] = account
	valTable[len + 2] = coins 
	valTable[len + 3] = nonce 

	Log(string.format("val = %v", valTable))
	ok, json = ToJson(valTable)
	Assert(ok)
	Assert(Put(userTableKey, json))

	Log("after put table")
	number = number + 1
	totalCoins = totalCoins + coins
	Assert(Put("user_number", number))
	Assert(Put("total_coins", totalCoins))
	Log(string.format("after put number, number = %d", number))

	if (number >= maxUserNumber)
	then
		Log("number enough")
		blockNumber = Height()
		ok, lastLuckyBlock = Get("last_lucky_block")
		Assert(ok)
		pHash = ParentHash()

		if (lastLuckyBlock < 0 or blockNumber - lastLuckyBlock >= 16 or blockNumber > lastLuckyBlock and pHash % 16 == 0)
		then
			Assert(Put("user_number", 0))
			Assert(Put("total_coins", 0))
			Assert(Put("last_lucky_block", blockNumber))
			Assert(getReward(blockNumber, totalCoins, number) == 0)
		end
	end

	return 0
end--f

--- getReward give reward to lucky dogs
-- @param_cnt 3
-- @return_cnt 1
-- @privilege private
function getReward(blockNumber, totalCoins, userNumber)
	print(string.format("get reward blockNumber = %d, coins = %f, user = %d", blockNumber, totalCoins, userNumber))
	Log(string.format("get reward blockNumber = %d, coins = %f, user = %d", blockNumber, totalCoins, userNumber))
	luckyNumber = blockNumber % 10
	ok, round = Get("round")
	Assert(ok)
	round = round + 1
	roundKey = string.format("round%d", round)
	roundValue = ""

	userTableKey = string.format("user_value%d", luckyNumber)
	ok, json = Get(userTableKey)
	Assert(ok)
	ok, valTable = ParseJson(json)
	Assert(ok)
	if (valTable == nil)
	then
		valTable = {}
	end

	res = {}
	res["UnwinUserList"] = {}
	for i = 0, 9, 1 do
		if (i ~= luckyNumber)
		then
			userTableKey = string.format("user_value%d", i)
			ok, json = Get(userTableKey)
			Assert(ok)
			ok, tmpTable = ParseJson(json)
			Assert(ok)

			tn = #tmpTable
			kn = math.floor((tn + 1) / 3)
			for j = 0, kn - 1, 1 do
				a0 = tmpTable[j * 3 + 1]
				a1 = tmpTable[j * 3 + 2]
				a2 = tmpTable[j * 3 + 3]
				unwinUser = {}
				unwinUser["Address"] = a0
				unwinUser["BetAmount"] = a1
				unwinUser["Amount"] = 0
				unwinUser["Nonce"] = a2
				unwinUser["LuckyNumber"] = i
				l0 = #(res["UnwinUserList"])
				res["UnwinUserList"][l0 + 1] = unwinUser
			end
		end
	end

	Assert(clearUserValue() == 0)

	totalCoins = totalCoins * 0.95
	totalVal = 0
	len = #valTable
	kNumber = math.floor((len + 1) / 3)
	Log(string.format("win len = ", len, ", key len = ", kNumber))

	for i = 0, kNumber - 1, 1 do
		totalVal = totalVal + valTable[i * 3 + 2]
	end
	print("totalval = ", totalVal)

	res["BlockHeight"] = blockNumber
	res["TotalUserNumber"] = userNumber
	res["WinUserNumber"] = kNumber
	res["TotalRewards"] = totalCoins
	res["WinUserList"] = {}

	for i = 0, kNumber - 1, 1 do
		address = valTable[i * 3 + 1]
		betCoins = valTable[i * 3 + 2]
		print("bet coins = ", betCoins)
		winUser = {}
		winUser["BetAmount"] = betCoins
		winUser["LuckyNumber"] = luckyNumber
		winCoins = betCoins / totalVal * totalCoins
		winUser["Address"] = address
		winUser["Amount"] = winCoins
		winUser["Nonce"] = valTable[i * 3 + 3]

		print("i  = ", i, " address = ", address, ", wincoins = ", winCoins)
		Assert(Withdraw(address, winCoins) == true)
		len = #(res["WinUserList"])
		res["WinUserList"][len + 1] = winUser
	end


	Log(string.format("res = %v", res))

	ok, roundValue = ToJson(res)
	Assert(ok)
	Assert(Put(roundKey, roundValue))
	Assert(Put("round", round))

	return 0
end--f

--- QueryUserNumber query user number now 
-- @param_cnt 0
-- @return_cnt 1
-- @privilege public
function QueryUserNumber()
	ok, r = Get("user_number")
	Assert(ok)
	return r
end--f

--- QueryTotalCoins query total coins
-- @param_cnt 0
-- @return_cnt 1
-- @privilege public
function QueryTotalCoins()
	ok, r = Get("total_coins")
	Assert(ok)
	return r
end--f

--- QueryLastLuckyBlock query last lucky block 
-- @param_cnt 0
-- @return_cnt 1
-- @privilege public
function QueryLastLuckyBlock()
	ok, r = Get("last_lucky_block")
	Assert(ok)
	return r
end--f

--- QueryMaxUserNumber query max user number 
-- @param_cnt 0
-- @return_cnt 1
-- @privilege public
function QueryMaxUserNumber()
	ok, r = Get("max_user_number")
	Assert(ok)
	return r
end--f

--- QueryRound query round
-- @param_cnt 0
-- @return_cnt 1
-- @privilege public
function QueryRound()
	ok, r = Get("round")
	Assert(ok)
	return r
end--f
`
	bdb, err := db.DatabaseFactory("redis")
	if err != nil {
		panic(err)
	}
	pdb := state.NewDatabase(bdb)
	pool := state.NewPool(pdb)

	pmain, _ := lua.NewDocCommentParser(main)
	pcall, _ := lua.NewDocCommentParser(buy)

	cmain, err := pmain.Parse()
	cmain.SetSender("publisher")
	if err != nil {
		panic(err)
	}
	ccall, err := pcall.Parse()
	ccall.SetSender("caller")
	if err != nil {
		panic(err)
	}

	//tmain := tx.NewTx(123, cmain)
	//tcall := tx.NewTx(456, ccall)

	pool.PutHM("iost", "publisher", state.MakeVFloat(100000000))
	pool.PutHM("iost", "caller", state.MakeVFloat(100000000))

	verifier := CacheVerifier{
		Verifier: Verifier{vmMonitor: newVMMonitor(), Context: vm.BaseContext()},
	}

	cmain.SetPrefix("main")
	ccall.SetPrefix("buy")

	verifier.StartVM(cmain)
	_, err = verifier.VerifyContract(cmain, pool)
	if err != nil {
		panic(err)
	}

	for i := 0; i < b.N; i++ {
		_, err = verifier.VerifyContract(ccall, pool)
		if err != nil {
			panic(err)
		}
	}

}
