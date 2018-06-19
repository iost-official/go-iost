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
	"fmt"
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
		txChan := make(chan message.Message, 100000)
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

		chConfirmBlock := make(chan *block.Block, 10000)
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

		Convey("concurrent", func() {
			txCnt := 100
			blockCnt := 1000
			bl := genBlocks(BlockCache, accountList, witnessList, blockCnt, txCnt, true)
			ch := make(chan int, 11)

			go func() {
				for _, blk := range bl {
					txPool.addBlockTx(blk)
				}
				ch <- 1
			}()

			txx := genTx(accountList[0], 10000)
			go func() {
				for i := 0; i < 10000; i++ {
					tx := genTx(accountList[0], 100000+i)
					txPool.addListTx(&tx)
				}
				ch <- 2
			}()

			go func() {
				for i := 0; i < 10000; i++ {
					tx := genTx(accountList[0], 1000000+i)
					broadTx := message.Message{
						Body:    tx.Encode(),
						ReqType: int32(network.ReqPublishTx),
					}
					txPool.AddTransaction(broadTx)
				}
				ch<-3
			}()
			////time.Sleep(5*time.Second)
			runCnt := 100
			go func() {
				for i:=0; i<runCnt ;i++  {
					//time.Sleep(1*time.Millisecond)
					txPool.BlockTxNum()
				}
				ch <- 4
			}()
			go func() {
				for i:=0; i<runCnt ;i++  {
					//time.Sleep(1*time.Millisecond)
					txPool.PendingTransactions(100000)
				}
				ch <- 5
			}()
			go func() {
				for i:=0; i<runCnt ;i++  {
					//time.Sleep(1*time.Millisecond)
					txPool.TransactionNum()
				}
				ch <- 6
			}()
			go func() {
				for i:=0; i<runCnt ;i++  {
					//time.Sleep(1*time.Millisecond)
					txPool.PendingTransactionNum()
				}
				ch <- 7
			}()
			go func() {
				for i:=0; i<runCnt ;i++  {
					//time.Sleep(1*time.Millisecond)
					txPool.Transaction(txx.TxID())

				}
				ch <- 8
			}()
			go func() {
				for i:=0; i<runCnt ;i++  {
					//time.Sleep(1*time.Millisecond)
					txPool.ExistTransaction(txx.TxID())
				}
				ch <- 9
			}()
			go func() {
				for i:=0; i<runCnt ;i++  {
					//time.Sleep(1*time.Millisecond)
					txPool.delTimeOutTx()
				}
				ch <- 10
			}()
			go func() {
				for i:=0; i<runCnt ;i++  {
					//time.Sleep(1*time.Millisecond)
					txPool.delTimeOutBlockTx()
				}
				ch <- 11
			}()

			for i := 0; i < 11; i++ {
				c:=<-ch
				fmt.Println("结束并发 i=", i, ", c=", c)
			}

		})

		Convey("addBlockTx", func() {
			txCnt := 20
			bl := genBlocks(BlockCache, accountList, witnessList, 2, txCnt, true)
			So(txPool.BlockTxNum(), ShouldEqual, 0)
			txPool.addBlockTx(bl[0])
			txPool.addBlockTx(bl[1])
			txPool.addBlockTx(bl[0])
			txPool.addBlockTx(bl[1])
			So(txPool.BlockTxNum(), ShouldEqual, 2)
			So(len(txPool.blockTx.blkTx[bl[0].HashID()].hashList), ShouldEqual, txCnt)
			So(len(txPool.blockTx.blkTx[bl[1].HashID()].hashList), ShouldEqual, txCnt)

			listTxCnt := 2
			for i := 0; i < listTxCnt; i++ {
				tx := genTx(accountList[0], 30+i)
				txPool.addListTx(&tx)
			}

			txPool.updatePending(txCnt)

			So(txPool.PendingTransactionNum(), ShouldEqual, listTxCnt)
			for _, tx := range bl[0].Content {
				//fmt.Println("add List tr hash:",tx.TxID(), " tr nonce:", tx.Nonce)
				txPool.addListTx(&tx)
			}

			So(txPool.TransactionNum(), ShouldEqual, len(bl[0].Content)+listTxCnt)
			//for hash, tx := range txPool.listTx.list {
			//	fmt.Println("print ListTx - tr hash:",hash, " nonoc:", tx.Nonce)
			//
			//}
			//
			//for hash, _ := range txPool.blockTx.blkTx[bl[0].TxID()].hashList {
			//	fmt.Println("0.hashList - tr hash:",hash)
			//
			//}
			txPool.checkIterateBlockHash.Add(bl[0].HashID())
			txPool.updatePending(500)

			//txList:=txPool.PendingTransactions()
			//for _,tx:=range txList {
			//	fmt.Println(tx.Nonce, ", tr hash:",tx.TxID())
			//}
			So(txPool.PendingTransactionNum(), ShouldEqual, listTxCnt)
			//So(txPool.PendingTransactionNum(), ShouldEqual, listTxCnt)
			//So(txPool.PendingTransactionNum(), ShouldEqual, listTxCnt)

		})
	})
}

func BenchmarkUpdatePending(b *testing.B) {
	BlockCache, accountList, witnessList, txPool := envInit(b)

	blockList := genBlocks(BlockCache, accountList, witnessList, 2, 500, true)

	listTxCnt := 500
	for i := 0; i < listTxCnt; i++ {
		tx := genTx(accountList[0], 30+i)
		txPool.addListTx(&tx)
	}

	txPool.addBlockTx(blockList[0])
	txPool.checkIterateBlockHash.Add(blockList[0].HashID())
	txPool.addBlockTx(blockList[1])
	txPool.checkIterateBlockHash.Add(blockList[1].HashID())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		txPool.updatePending(500)
	}

}

func envInit(b *testing.B) (blockcache.BlockCache, []account.Account, []string, *TxPoolServer) {
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
	mockCtr := NewController(b)
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

	chConfirmBlock := make(chan *block.Block, 10)
	txPool, err := NewTxPoolServer(BlockCache, chConfirmBlock)
	if err != nil {
		panic("NewTxPoolServer error")
	}
	txPool.Start()
	return BlockCache, accountList, witnessList, txPool
}

func genTx(a account.Account, nonce int) tx.Tx {
	main := lua.NewMethod(2, "main", 0, 1)
	code := `function main()
				Put("hello", "world")
				return "success"
			end`
	lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Publisher: vm.IOSTAccount(a.ID)}, code, main)

	_tx := tx.NewTx(int64(nonce), &lc)
	//_tx, _ = tx.SignTx(_tx, a)
	return _tx
}

func genBlocks(p blockcache.BlockCache, accountList []account.Account, witnessList []string, blockCnt int, txCnt int, continuity bool) (blockPool []*block.Block) {

	//blockLen := p.blockCache.ConfirmedLength()
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
			Info:       []byte(""),
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
