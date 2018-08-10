package new_txpool

import (
	. "github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/consensus/common"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/message"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/log"
	"github.com/iost-official/Go-IOS-Protocol/network"
	"github.com/iost-official/Go-IOS-Protocol/network/mocks"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestNewTxPoolImpl(t *testing.T) {
	Convey("test NewTxPoolServer", t, func() {
		var accountList []account.Account
		var witnessList []string

		acc := common.Base58Decode("3BZ3HWs2nWucCCvLp7FRFv1K7RR3fAjjEQccf9EJrTv4")
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

		config := &p2p.Config{
			ListenAddr: "0.0.0.0:8088",
		}

		node, err := p2p.NewNetService(config)

		log.NewLogger("iost")

		conf := &common.Config{}

		gl, err := global.New(conf)
		So(err, ShouldBeNil)

		BlockCache, err := blockcache.NewBlockCache(gl)
		So(err, ShouldBeNil)

		txPool, err := NewTxPoolImpl(gl, BlockCache, node)
		So(err, ShouldBeNil)

		txPool.Start()

		Convey("AddTx", func() {

			tx := genTxMsg(accountList[0], expiration)

			So(txPool.testPendingTxsNum(), ShouldEqual, 0)
			txPool.AddTx(tx)
			So(txPool.testPendingTxsNum(), ShouldEqual, 1)
		})

		Convey("txTimeOut", func() {

			tx := genTx(accountList[0], expiration*2)

			tx.Time -= int64(expiration*1e9 + 1*1e9)
			b := txPool.txTimeOut(tx)
			So(b, ShouldBeTrue)

			tx = genTx(accountList[0], expiration)

			tx.Expiration -= int64(expiration * 1e9 * 3)
			b = txPool.txTimeOut(tx)
			So(b, ShouldBeTrue)

		})
		//
		//Convey("delTimeOutTx", func() {
		//
		//	tx := genTx(accountList[0], 1)
		//	So(txPool.TransactionNum(), ShouldEqual, 0)
		//
		//	tx.Time -= int64(expiration*1e9 + 1*1e9)
		//	txPool.addListTx(&tx)
		//	So(txPool.TransactionNum(), ShouldEqual, 1)
		//
		//	txPool.delTimeOutTx()
		//	So(txPool.TransactionNum(), ShouldEqual, 0)
		//
		//})
		//
		//Convey("concurrent", func() {
		//	txCnt := 100
		//	blockCnt := 1000
		//	bl := genBlocks(BlockCache, accountList, witnessList, blockCnt, txCnt, true)
		//	ch := make(chan int, 11)
		//
		//	go func() {
		//		for _, blk := range bl {
		//			txPool.addBlockTx(blk)
		//		}
		//		ch <- 1
		//	}()
		//
		//	txx := genTx(accountList[0], 10000)
		//	go func() {
		//		for i := 0; i < 10000; i++ {
		//			tx := genTx(accountList[0], 100000+i)
		//			txPool.addListTx(&tx)
		//		}
		//		ch <- 2
		//	}()
		//
		//	go func() {
		//		for i := 0; i < 10000; i++ {
		//			tx := genTx(accountList[0], 1000000+i)
		//			broadTx := message.Message{
		//				Body:    tx.Encode(),
		//				ReqType: int32(network.ReqPublishTx),
		//			}
		//			txPool.AddTransaction(&broadTx)
		//		}
		//		ch <- 3
		//	}()
		//	////time.Sleep(5*time.Second)
		//	runCnt := 100
		//	go func() {
		//		for i := 0; i < runCnt; i++ {
		//			//time.Sleep(1*time.Millisecond)
		//			txPool.BlockTxNum()
		//		}
		//		ch <- 4
		//	}()
		//	go func() {
		//		for i := 0; i < runCnt; i++ {
		//			//time.Sleep(1*time.Millisecond)
		//			txPool.PendingTransactions(100000)
		//		}
		//		ch <- 5
		//	}()
		//	go func() {
		//		for i := 0; i < runCnt; i++ {
		//			//time.Sleep(1*time.Millisecond)
		//			txPool.TransactionNum()
		//		}
		//		ch <- 6
		//	}()
		//	go func() {
		//		for i := 0; i < runCnt; i++ {
		//			//time.Sleep(1*time.Millisecond)
		//			txPool.PendingTransactionNum()
		//		}
		//		ch <- 7
		//	}()
		//	go func() {
		//		for i := 0; i < runCnt; i++ {
		//			//time.Sleep(1*time.Millisecond)
		//			txPool.Transaction(txx.TxID())
		//
		//		}
		//		ch <- 8
		//	}()
		//	go func() {
		//		for i := 0; i < runCnt; i++ {
		//			//time.Sleep(1*time.Millisecond)
		//			txPool.ExistTransaction(txx.TxID())
		//		}
		//		ch <- 9
		//	}()
		//	go func() {
		//		for i := 0; i < runCnt; i++ {
		//			//time.Sleep(1*time.Millisecond)
		//			txPool.delTimeOutTx()
		//		}
		//		ch <- 10
		//	}()
		//	go func() {
		//		for i := 0; i < runCnt; i++ {
		//			//time.Sleep(1*time.Millisecond)
		//			txPool.delTimeOutBlockTx()
		//		}
		//		ch <- 11
		//	}()
		//
		//	for i := 0; i < 11; i++ {
		//		<-ch
		//	}
		//
		//})
		//
		//Convey("addBlockTx", func() {
		//	txCnt := 20
		//	bl := genBlocks(BlockCache, accountList, witnessList, 2, txCnt, true)
		//	So(txPool.BlockTxNum(), ShouldEqual, 0)
		//	txPool.addBlockTx(bl[0])
		//	txPool.addBlockTx(bl[1])
		//	txPool.addBlockTx(bl[0])
		//	txPool.addBlockTx(bl[1])
		//	So(txPool.BlockTxNum(), ShouldEqual, 2)
		//	So(len(txPool.blockTx.blkTx[bl[0].HashID()].hashList), ShouldEqual, txCnt)
		//	So(len(txPool.blockTx.blkTx[bl[1].HashID()].hashList), ShouldEqual, txCnt)
		//
		//	listTxCnt := 2
		//	for i := 0; i < listTxCnt; i++ {
		//		tx := genTx(accountList[0], 30+i)
		//		txPool.addListTx(&tx)
		//	}
		//
		//	txPool.updatePending(txCnt)
		//
		//	So(txPool.PendingTransactionNum(), ShouldEqual, listTxCnt)
		//	for _, tx := range bl[0].Content {
		//		//fmt.Println("add List tr hash:",tx.TxID(), " tr nonce:", tx.Nonce)
		//		txPool.addListTx(&tx)
		//	}
		//
		//	So(txPool.TransactionNum(), ShouldEqual, len(bl[0].Content)+listTxCnt)
		//	txPool.checkIterateBlockHash.Add(bl[0].HashID())
		//	txPool.updatePending(500)
		//
		//	So(txPool.PendingTransactionNum(), ShouldEqual, listTxCnt)
		//
		//})
	})
}

//func BenchmarkUpdatePending(b *testing.B) {
//	BlockCache, accountList, witnessList, txPool := envInit(b)
//
//	blockList := genBlocks(BlockCache, accountList, witnessList, 2, 500, true)
//
//	listTxCnt := 500
//	for i := 0; i < listTxCnt; i++ {
//		tx := genTx(accountList[0], 30+i)
//		txPool.addListTx(&tx)
//	}
//
//	txPool.addBlockTx(blockList[0])
//	txPool.longestChainHash.Add(blockList[0].HashID())
//	txPool.addBlockTx(blockList[1])
//	txPool.longestChainHash.Add(blockList[1].HashID())
//
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		txPool.updatePending(500)
//	}
//
//}

func envInit(b *testing.B) (blockcache.BlockCache, []account.Account, []string, *TxPoolImpl) {
	var accountList []account.Account
	var witnessList []string

	acc := common.Base58Decode("3BZ3HWs2nWucCCvLp7FRFv1K7RR3fAjjEQccf9EJrTv4")
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
	txChan := make(chan message.Message, 100000)
	mockRouter.EXPECT().FilteredChan(Any()).Return(txChan, nil)

	log.NewLogger("iost")

	conf := &common.Config{}

	gl, err := global.New(conf)
	So(err, ShouldBeNil)

	BlockCache, err := blockcache.NewBlockCache(gl)
	So(err, ShouldBeNil)

	txPool, err := NewTxPoolImpl(BlockCache, network.Route, gl)
	So(err, ShouldBeNil)

	txPool.Start()

	return BlockCache, accountList, witnessList, txPool
}

func genTx(a account.Account, expirationIter int64) *tx.Tx {
	actions := []tx.Action{}
	actions = append(actions, tx.Action{
		Contract:   "contract1",
		ActionName: "actionname1",
		Data:       "{\"num\": 1, \"message\": \"contract1\"}",
	})
	actions = append(actions, tx.Action{
		Contract:   "contract2",
		ActionName: "actionname2",
		Data:       "1",
	})

	ex := time.Now().UnixNano()
	ex += expirationIter * 1e9

	tx := tx.NewTx(actions, [][]byte{a.Pubkey}, 100000, 100, ex)

	return &tx
}

func genTxMsg(a account.Account, expirationIter int64) p2p.IncomingMessage {
	t := genTx(a, expirationIter)

	broadTx := p2p.IncomingMessage{
		data: t.Encode(),
		typ:  int32(p2p.ReqPublishTx),
	}

	return broadTx
}

func genBlocks(p blockcache.BlockCache, accountList []account.Account, witnessList []string, blockCnt int, txCnt int, continuity bool) (blockPool []*block.Block) {

	slot := consensus_common.GetCurrentTimestamp().Slot

	for i := 0; i < blockCnt; i++ {
		var hash []byte

		if continuity == false {
			hash[i%len(hash)] = byte(i % 256)
		}
		blk := block.Block{Txs: []*tx.Tx{}, Head: block.BlockHead{
			Version:    0,
			ParentHash: hash,
			MerkleHash: make([]byte, 0),
			Info:       []byte(""),
			Number:     int64(i + 1),
			Witness:    witnessList[0],
			Time:       slot + int64(i),
		}}

		for i := 0; i < txCnt; i++ {
			blk.Txs = append(blk.Txs, genTx(accountList[0], int64(i)))
		}
		blockPool = append(blockPool, &blk)
	}
	return
}
