package tx

import (
	"testing"
	//	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"

	"fmt"

	"github.com/iost-official/prototype/vm/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func gentx() Tx {
	main := lua.NewMethod(0, "main", 0, 1)
	code := `function main()
				Put("hello", "world")
				return "success"
			end`
	lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Publisher: vm.IOSTAccount("ahaha")}, code, main)

	return NewTx(int64(0), &lc)
}
func TestTx(t *testing.T) {
	Convey("Test of Tx", t, func() {
		ctl := gomock.NewController(t)

		mockContract := vm_mock.NewMockContract(ctl)
		mockContract.EXPECT().Encode().AnyTimes().Return([]byte{1, 2, 3})

		a1, _ := account.NewAccount(nil)
		a2, _ := account.NewAccount(nil)
		a3, _ := account.NewAccount(nil)

		Convey("sign and verify", func() {
			tx := NewTx(int64(0), mockContract)
			sig1, err := SignContract(tx, a1)

			So(tx.VerifySigner(sig1), ShouldBeTrue)

			tx.Signs = append(tx.Signs, sig1)
			sig2, err := SignContract(tx, a2)
			tx.Signs = append(tx.Signs, sig2)

			err = tx.VerifySelf()
			So(err.Error(), ShouldEqual, "publisher error")

			tx3, err := SignTx(tx, a3)
			So(err, ShouldBeNil)
			err = tx3.VerifySelf()
			So(err, ShouldBeNil)

			tx.Signs[0] = common.Signature{
				Algorithm: common.Secp256k1,
				Sig:       []byte("hello"),
				Pubkey:    []byte("world"),
			}
			err = tx.VerifySelf()
			So(err.Error(), ShouldEqual, "signer error")
		})

	})
}

func TestProblemTxs(t *testing.T) {
	errlua := `--- main 一元夺宝
-- snatch treasure with 1 coin !
-- @gas_limit 100000
-- @gas_price 0.01
-- @param_cnt 0
-- @return_cnt 1
-- @publisher walleta
function main()
	Put("max_user_number", 20)
	Put("user_number", 0)
	Put("winner", "")
	Put("claimed", "false")
    return "success"
end--f

--- BuyCoin buy coins
-- buy some coins
-- @param_cnt 2
-- @return_cnt 1
function BuyCoin(account, buyNumber)
	if (buyNumber <= 0)
	then
	    return "buy number should be more than zero"
	end

	maxUserNumber = Get("max_user_number")
    number = Get("user_number")
	if (number >= maxUserNumber or number + buyNumber > maxUserNumber)
	then
	    return string.format("max user number exceed, only %d coins left", maxUserNumber - number)
	end

	-- print(string.format("deposit account = %s, number = %d", account, buyNumber))
	Deposit(account, buyNumber)

	win = false
	for i = 0, buyNumber - 1, 1 do
	    win = win or winAfterBuyOne(number)
	    number = number + 1
	end
	Put("user_number", number)

	if (win)
	then
	    Put("winner", account)
	end

    return "success"
end--f

--- winAfterBuyOne win after buy one
-- @param_cnt 1
-- @return_cnt 1
function winAfterBuyOne(number)
	win = Random(1 - 1.0 / (number + 1))
	return win
end--f

--- QueryWinner query winner
-- @param_cnt 0
-- @return_cnt 1
function QueryWinner()
	return Get("winner")
end--f

--- QueryClaimed query claimed
-- @param_cnt 0
-- @return_cnt 1
function QueryClaimed()
	return Get("claimed")
end--f

--- QueryUserNumber query user number 
-- @param_cnt 0
-- @return_cnt 1
function QueryUserNumber()
	return Get("user_number")
end--f

--- QueryMaxUserNumber query max user number 
-- @param_cnt 0
-- @return_cnt 1
function QueryMaxUserNumber()
	return Get("max_user_number")
end--f

--- Claim claim prize
-- @param_cnt 0
-- @return_cnt 1
function Claim()
	claimed = Get("claimed")
	if (claimed == "true")
	then
		return "price has been claimed"
	end
	number = Get("user_number")
	maxUserNumber = Get("max_user_number")
	if (number < maxUserNumber)
	then
		return string.format("game not end yet! user_number = %d, max_user_number = %d", number, maxUserNumber)
	end
	winner = Get("winner")

	Put("claimed", "true")

	Withdraw(winner, number)
	return "success"
end--f
`
	commonlua := `--- main
-- 输出hello world
-- @gas_limit 10000
-- @gas_price 0.0001
-- @param_cnt 0
-- @return_cnt 1
-- @publisher walleta
function main()
    print(ParentHash())
    print(Height())
    print(Witness())
    print(Now())
end--f

`

	syntaxlua := `
--- main 买1元
-- snatch treasure with 1 coin
-- @gas_limit 1000000
-- @gas_price 0.0001
-- @param_cnt 0
-- @return_cnt 1
-- @publisher walleta
function main()
    return 15 != 20;
end--f`

	Convey("test 1", t, func() {
		_ = commonlua
		_ = errlua
		_ = syntaxlua

		parser, err := lua.NewDocCommentParser(syntaxlua)

		So(err, ShouldBeNil)

		contract, err := parser.Parse()

		mTx := NewTx(int64(123), contract)

		//bytes := mTx.Encode()

		seckey := common.Base58Decode("3BZ3HWs2nWucCCvLp7FRFv1K7RR3fAjjEQccf9EJrTv4")
		//fmt.Println(common.Base58Encode(seckey))
		acc, err := account.NewAccount(seckey)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		stx, err := SignTx(mTx, acc)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		So(common.Base58Encode(mTx.publishHash()), ShouldEqual, common.Base58Encode(stx.publishHash()))
		bytes := stx.Encode()
		var ptx, p2tx Tx
		err = ptx.Decode(bytes)
		So(err, ShouldBeNil)

		buf2 := ptx.Encode()
		p2tx.Decode(buf2)

		So(stx.Contract.Code(), ShouldEqual, ptx.Contract.Code())
		So(p2tx.Contract.Info().Prefix, ShouldEqual, ptx.Contract.Info().Prefix)

		So(common.Base58Encode(ptx.publishHash()), ShouldEqual, common.Base58Encode(stx.publishHash()))

		err = ptx.VerifySelf()
		So(err, ShouldBeNil)
	})
}
