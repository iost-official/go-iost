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
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/core/blockcache"
	"github.com/iost-official/prototype/log"
	"github.com/iost-official/prototype/vm/lua"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/consensus/common"
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

		log.NewLogger("iost")

		BlockCache := blockcache.NewBlockCache(blockChain, state.StdPool, len(witnessList)*2/3)

		chConfirmBlock := make(chan block.Block, 10)
		txPool, err := NewTxPoolServer(BlockCache, chConfirmBlock)
		So(err, ShouldBeNil)

		txPool.Start()

		Convey("addListTx", func() {

			tx := genTx(accountList[0], 1)
			So(txPool.TransactionNum(), ShouldEqual, 0)
			txPool.addListTx(&tx)
			So(txPool.TransactionNum(), ShouldEqual, 1)

		})

		Convey("txTimeOut", func() {

			tx := genTx(accountList[0], 1)

			tx.Time -= int64(filterTime*1e9 + 1*1e9)
			b := txPool.txTimeOut(&tx)
			So(b, ShouldBeTrue)

		})

		Convey("delTimeOutTx", func() {

			tx := genTx(accountList[0], 1)
			So(txPool.TransactionNum(), ShouldEqual, 0)

			tx.Time -= int64(filterTime*1e9 + 1*1e9)
			txPool.addListTx(&tx)
			So(txPool.TransactionNum(), ShouldEqual, 1)

			txPool.delTimeOutTx()
			So(txPool.TransactionNum(), ShouldEqual, 0)

		})

		Convey("addBlockTx", func() {
			bl := genBlocks(BlockCache, accountList, witnessList, 10, 20, true)
			So(txPool.BlockTxNum(), ShouldEqual, 0)
			txPool.addBlockTx(bl[0])
			So(txPool.BlockTxNum(), ShouldEqual, 1)

		})
	})
}

func genTx(a account.Account, nonce int) tx.Tx {
	main := lua.NewMethod(2, "main", 0, 1)
	code := `function main()
				Put("hello", "world")
				return "success"
			end`
	lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Publisher: vm.IOSTAccount(a.ID)}, code, main)

	_tx := tx.NewTx(int64(nonce), &lc)
	_tx, _ = tx.SignTx(_tx, a)
	return _tx
}

func genBlocks(p blockcache.BlockCache, accountList []account.Account, witnessList []string, blockCnt int, txCnt int, continuity bool) (blockPool []*block.Block) {

	//blockLen := p.BlockCache.ConfirmedLength()
	//fmt.Println(blockLen)

	//blockNum := 1000
	slot := consensus_common.GetCurrentTimestamp().Slot

	for i := 0; i < blockCnt; i++ {
		var hash []byte

		//make every block has no parent
		if continuity == false {
			hash[i%len(hash)] = byte(i % 256)
		}
		blk := block.Block{Content: []tx.Tx{}, Head: block.BlockHead{
			Version:    0,
			ParentHash: hash,
			TreeHash:   make([]byte, 0),
			BlockHash:  make([]byte, 0),
			Info:       []byte("test"),
			Number:     int64(i + 1),
			Witness:    witnessList[0],
			Time:       slot + int64(i),
		}}

		for i := 0; i < txCnt; i++ {
			blk.Content = append(blk.Content, genTx(accountList[0], i))
		}
		blockPool = append(blockPool, &blk)
	}
	return
}
