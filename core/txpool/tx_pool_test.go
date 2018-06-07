package txpool

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/network"
	. "github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/network/mocks"
	"github.com/iost-official/prototype/core/message"
	"fmt"
	"github.com/iost-official/prototype/core/tx"
)

func TestNewTxPoolServer(t *testing.T) {
	Convey("test NewTxPoolServer", t, func() {
		var accountList []account.Account
		var witnessList []string

		acc := common.Base58Decode("BRpwCKmVJiTTrPFi6igcSgvuzSiySd7Exxj7LGfqieW9")
		_account, err := account.NewAccount(acc)
		if err != nil {
			panic("account.NewAccount error")
		}
		accountList = append(accountList, _account)
		witnessList = append(witnessList, _account.ID)
		//_accId := _account.ID

		for i := 1; i < 3; i++ {
			_account, err := account.NewAccount(nil)
			if err != nil {
				panic("account.NewAccount error")
			}
			accountList = append(accountList, _account)
			witnessList = append(witnessList, _account.ID)
		}

		tx.LdbPath = ""

		mockCtr := NewController(t)
		mockRouter := protocol_mock.NewMockRouter(mockCtr)

		network.Route = mockRouter
		// 设置第一个通道txchan
		txChan := make(chan message.Message, 1)
		mockRouter.EXPECT().FilteredChan(Any()).Return(txChan, nil)

		txDb := tx.TxDbInstance()
		if txDb == nil {
			panic("txDB error")
		}

		blockChain, err := block.Instance()
		if err != nil {
			panic("block.Instance error")
		}

		err = state.PoolInstance()
		if err != nil {
			panic("state.PoolInstance error")
		}

		BlockCache := block.NewBlockCache(blockChain, state.StdPool, len(witnessList)*2/3)

		chConfirmBlock := make(chan block.Block, 10)
		fmt.Println(BlockCache, chConfirmBlock)
		txPool, err:=NewTxPoolServer(BlockCache, chConfirmBlock)
		So(err, ShouldBeTrue)

		txPool.Start()


	})
}