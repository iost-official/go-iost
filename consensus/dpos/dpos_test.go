package dpos

import (
	"fmt"
	"testing"

	. "github.com/bouk/monkey"
	. "github.com/golang/mock/gomock"

	"sync"
	"time"

	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/consensus/common"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/core/mocks"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/network"
	"github.com/iost-official/prototype/network/mocks"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewDPoS(t *testing.T) {
	Convey("Test of NewDPoS", t, func() {
		mockCtr := NewController(t)
		mockRouter := protocol_mock.NewMockRouter(mockCtr)
		mockBc := core_mock.NewMockChain(mockCtr)
		mockPool := core_mock.NewMockPool(mockCtr)
		mockPool.EXPECT().Copy().Return(mockPool).AnyTimes()
		mockPool.EXPECT().PutHM(Any(), Any(), Any()).AnyTimes().Return(nil)
		mockPool.EXPECT().Flush().AnyTimes().Return(nil)

		network.Route = mockRouter
		//获取router实例
		guard := Patch(network.RouterFactory, func(_ string) (network.Router, error) {
			return mockRouter, nil
		})
		defer guard.Unpatch()

		heightChan := make(chan message.Message, 1)
		blkChan := make(chan message.Message, 1)
		mockRouter.EXPECT().FilteredChan(Any()).Return(heightChan, nil)
		mockRouter.EXPECT().FilteredChan(Any()).Return(blkChan, nil)

		// 设置第一个通道txchan
		txChan := make(chan message.Message, 1)
		mockRouter.EXPECT().FilteredChan(Any()).Return(txChan, nil)

		// 设置第二个通道Blockchan
		blockChan := make(chan message.Message, 1)
		mockRouter.EXPECT().FilteredChan(Any()).Return(blockChan, nil)

		blk := block.Block{Content: []tx.Tx{}, Head: block.BlockHead{
			Version:    0,
			ParentHash: []byte("111"),
			TreeHash:   make([]byte, 0),
			BlockHash:  make([]byte, 0),
			Info:       []byte("test"),
			Number:     int64(1),
			Witness:    "11111",
			Time:       1111,
		}}

		// 创世块的询问和插入
		var getNumber uint64
		var pushNumber int64
		mockBc.EXPECT().GetBlockByNumber(Any()).Return(nil).AnyTimes()
		//mockBc.EXPECT().GetBlockByNumber(Any()).AnyTimes().Return(&blk)
		//	Do(func(number uint64) *block.Block {
		//	getNumber = number
		//	return &blk
		//})
		mockBc.EXPECT().Length().AnyTimes().Do(func() uint64 { var r uint64 = 0; return r })
		mockBc.EXPECT().Top().AnyTimes().Return(&blk)
		mockBc.EXPECT().Push(Any()).Do(func(block *block.Block) error {
			pushNumber = block.Head.Number
			return nil
		})

		p, _ := NewDPoS(
			account.Account{
				ID:     "id0",
				Pubkey: []byte("pubkey"),
				Seckey: []byte("seckey"),
			},
			mockBc,
			mockPool,
			[]string{},
		)

		So(p.Account.GetId(), ShouldEqual, "id0")
		So(getNumber, ShouldEqual, 0)
		So(pushNumber, ShouldEqual, 0)
	})

}

func envinit(t *testing.T) (*DPoS, []account.Account, []string) {
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
	//获取router实例
	guard := Patch(network.RouterFactory, func(_ string) (network.Router, error) {
		return mockRouter, nil
	})

	heightChan := make(chan message.Message, 1)
	blkSyncChan := make(chan message.Message, 1)
	mockRouter.EXPECT().FilteredChan(Any()).Return(heightChan, nil)
	mockRouter.EXPECT().FilteredChan(Any()).Return(blkSyncChan, nil)

	//设置第一个通道txchan
	txChan := make(chan message.Message, 1)
	mockRouter.EXPECT().FilteredChan(Any()).Return(txChan, nil)

	//设置第二个通道Blockchan
	blkChan := make(chan message.Message, 1)
	mockRouter.EXPECT().FilteredChan(Any()).Return(blkChan, nil)

	defer guard.Unpatch()

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

	//verifyFunc := func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error) {
	//	return pool, nil
	//}

	//blockCache := consensus_common.NewBlockCache(blockChain, state.StdPool, len(witnessList)*2/3)
	//seckey := common.Sha256([]byte("SeckeyId0"))
	//pubkey := common.CalcPubkeyInSecp256k1(seckey)
	p, err := NewDPoS(accountList[0], blockChain, state.StdPool, witnessList)
	if err != nil {
		t.Errorf("NewDPoS error")
	}

	return p, accountList, witnessList
}

//func TestRunGenerateBlock(t *testing.T) {
//	Convey("Test of Run (Generate Block)", t, func() {
//		p, _, _ := envinit(t)
//		_tx := genTx(p, 998)
//		p.blockCache.AddTx(&_tx)
//		bc := p.blockCache.LongestChain()
//		pool := p.blockCache.LongestPool()
//		blk := p.genBlock(p.account, bc, pool)
//		So(len(blk.Content), ShouldEqual, 1)
//		So(blk.Content[0].Nonce, ShouldEqual, 998)
//		p.blockCache.Draw()
//	})
//}
//
////clear BlockDB and TxDB files before running this test
//func TestRunConfirmBlock(t *testing.T) {
//	Convey("Test of Run ConfirmBlock", t, func() {
//		cmd := exec.Command("/bin/bash", "-c", "cd $GOPATH/src/github.com/iost-official/prototype/consensus/dpos && rm -rf blockDB txDB")
//		cmd.Run()
//
//		p, accList, witnessList := envinit(t)
//		_tx := genTx(p, 998)
//		p.blockCache.AddTx(&_tx)
//
//		for i := 0; i < 5; i++ {
//			wit := ""
//			for wit != witnessList[0] {
//				currentTimestamp := consensus_common.GetCurrentTimestamp()
//				wit = witnessOfTime(&p.globalStaticProperty, &p.globalDynamicProperty, currentTimestamp)
//			}
//
//			bc := p.blockCache.LongestChain()
//			pool := p.blockCache.LongestPool()
//
//			blk := p.genBlock(p.account, bc, pool)
//			p.blockCache.ResetTxPoool()
//			p.globalDynamicProperty.update(&blk.Head)
//			err := p.blockCache.Add(blk, p.blockVerify)
//			fmt.Println(err)
//		}
//
//		So(p.blockCache.ConfirmedLength(), ShouldEqual, 1)
//		for i := 1; i < 3; i++ {
//			wit := ""
//			for wit != witnessList[i] {
//				currentTimestamp := consensus_common.GetCurrentTimestamp()
//				wit = witnessOfTime(&p.globalStaticProperty, &p.globalDynamicProperty, currentTimestamp)
//			}
//
//			bc := p.blockCache.LongestChain()
//			pool := p.blockCache.LongestPool()
//			blk := p.genBlock(accList[i], bc, pool)
//			p.blockCache.ResetTxPoool()
//			p.globalDynamicProperty.update(&blk.Head)
//			/*
//				guard := Patch(witnessOfTime, func(_ *globalStaticProperty, _ *globalDynamicProperty, _ consensus_common.Timestamp) string {
//					return witnessList[i]
//				})
//				defer guard.Unpatch()
//			*/
//			err := p.blockCache.Add(blk, p.blockVerify)
//			fmt.Println(err)
//			if i == 1 {
//				So(p.blockCache.ConfirmedLength(), ShouldEqual, 1)
//			}
//			if i == 2 {
//				So(p.blockCache.ConfirmedLength(), ShouldEqual, 6)
//			}
//		}
//
//		p.blockCache.Draw()
//	})
//}
//
////this need to be checked again
//func TestRunMultipleBlocks(t *testing.T) {
//	Convey("Test of Run (Multiple Blocks)", t, func() {
//		p, _, witnessList := envinit(t)
//		_tx := genTx(p, 998)
//		p.blockCache.AddTx(&_tx)
//
//		bc := p.blockCache.LongestChain()
//		pool := p.blockCache.LongestPool()
//
//		for i := 100; i < 105; i++ {
//			if i == 103 {
//				bc = p.blockCache.LongestChain()
//				pool = p.blockCache.LongestPool()
//			}
//			wit := ""
//			for wit != witnessList[0] {
//
//				currentTimestamp := consensus_common.GetCurrentTimestamp()
//				wit = witnessOfTime(&p.globalStaticProperty, &p.globalDynamicProperty, currentTimestamp)
//			}
//
//			blk := p.genBlock(p.account, bc, pool)
//			p.blockCache.ResetTxPoool()
//			p.globalDynamicProperty.update(&blk.Head)
//
//			blk.Head.Time = int64(i)
//
//			headInfo := generateHeadInfo(blk.Head)
//			sig, _ := common.Sign(common.Secp256k1, headInfo, p.account.Seckey)
//			blk.Head.Signature = sig.Encode()
//
//			err := p.blockCache.Add(blk, p.blockVerify)
//			fmt.Println(err)
//		}
//		p.blockCache.Draw()
//	})
//}
func generateTestBlockMsg(witness string, secKeyRaw string, number int64, parentHash []byte) (block.Block, message.Message) {
	blk := block.Block{
		Head: block.BlockHead{
			Number:     number,
			ParentHash: parentHash,
			Witness:    witness,
			Time:       consensus_common.GetCurrentTimestamp().Slot,
		},
	}
	headInfo := generateHeadInfo(blk.Head)
	sig, _ := common.Sign(common.Secp256k1, headInfo, common.Sha256([]byte(secKeyRaw)))
	blk.Head.Signature = sig.Encode()
	msg := message.Message{
		Time:    time.Now().Unix(),
		From:    "",
		To:      "",
		ReqType: int32(network.ReqNewBlock),
		Body:    blk.Encode(),
	}
	return blk, msg
}

//go test -bench=. -benchmem -run=nonce
func BenchmarkAddBlockCache(b *testing.B) {
	//benchAddBlockCache(b,10,true)
	benchAddBlockCache(b, 10, false)
}

/*
func BenchmarkGetBlock(b *testing.B) {
//	benchGetBlock(b,10,true)
	benchGetBlock(b,10,false)
}
*/
func BenchmarkBlockVerifier(b *testing.B) { benchBlockVerifier(b) }
func BenchmarkTxCache(b *testing.B) {
	//benchTxCache(b,true)
	benchTxCache(b, true)
}

/*
func BenchmarkTxCachePara(b *testing.B) {
	benchTxCachePara(b)
}
*/
/*
func BenchmarkTxDb(b *testing.B) {
	//benchTxDb(b,true)
	benchTxDb(b,false)
}
*/
func BenchmarkBlockHead(b *testing.B) { benchBlockHead(b) }
func BenchmarkGenerateBlock(b *testing.B) {
	benchGenerateBlock(b, 6000)
}

func envInit(b *testing.B) (*DPoS, []account.Account, []string) {
	var accountList []account.Account
	var witnessList []string

	acc := common.Base58Decode("BRpwCKmVJiTTrPFi6igcSgvuzSiySd7Exxj7LGfqieW9")
	_account, err := account.NewAccount(acc)
	if err != nil {
		panic("account.NewAccount error")
	}
	accountList = append(accountList, _account)
	witnessList = append(witnessList, _account.ID)
	_accId := _account.ID

	for i := 1; i < 3; i++ {
		_account, err := account.NewAccount(nil)
		if err != nil {
			panic("account.NewAccount error")
		}
		accountList = append(accountList, _account)
		witnessList = append(witnessList, _accId)
	}

	tx.LdbPath = ""

	mockCtr := NewController(b)
	mockRouter := protocol_mock.NewMockRouter(mockCtr)

	network.Route = mockRouter
	//获取router实例
	guard := Patch(network.RouterFactory, func(_ string) (network.Router, error) {
		return mockRouter, nil
	})

	heightChan := make(chan message.Message, 1)
	blkSyncChan := make(chan message.Message, 1)
	mockRouter.EXPECT().FilteredChan(Any()).Return(heightChan, nil)
	mockRouter.EXPECT().FilteredChan(Any()).Return(blkSyncChan, nil)

	//设置第一个通道txchan
	txChan := make(chan message.Message, 1)
	mockRouter.EXPECT().FilteredChan(Any()).Return(txChan, nil)

	//设置第二个通道Blockchan
	blkChan := make(chan message.Message, 1)
	mockRouter.EXPECT().FilteredChan(Any()).Return(blkChan, nil)

	defer guard.Unpatch()

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

	//verifyFunc := func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error) {
	//	return pool, nil
	//}

	//blockCache := consensus_common.NewBlockCache(blockChain, state.StdPool, len(witnessList)*2/3)
	//seckey := common.Sha256([]byte("SeckeyId0"))
	//pubkey := common.CalcPubkeyInSecp256k1(seckey)
	p, err := NewDPoS(accountList[0], blockChain, state.StdPool, witnessList)
	if err != nil {
		b.Errorf("NewDPoS error")
	}
	return p, accountList, witnessList
}

func genTx(p *DPoS, nonce int) tx.Tx {
	main := lua.NewMethod(2, "main", 0, 1)
	code := `function main()
				Put("hello", "world")
				return "success"
			end`
	lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Publisher: vm.IOSTAccount(p.account.ID)}, code, main)

	_tx := tx.NewTx(int64(nonce), &lc)
	_tx, _ = tx.SignTx(_tx, p.account)
	return _tx
}

func genBlockHead(p *DPoS) {
	blk := block.Block{Content: []tx.Tx{}, Head: block.BlockHead{
		Version:    0,
		ParentHash: nil,
		TreeHash:   make([]byte, 0),
		BlockHash:  make([]byte, 0),
		Info:       []byte("test"),
		Number:     int64(1),
		Witness:    p.account.ID,
		Time:       int64(0),
	}}

	headInfo := generateHeadInfo(blk.Head)
	sig, _ := common.Sign(common.Secp256k1, headInfo, p.account.Seckey)
	blk.Head.Signature = sig.Encode()
}

func genBlocks(p *DPoS, accountList []account.Account, witnessList []string, n int, txCnt int, continuity bool) (blockPool []*block.Block) {
	confChain := p.blockCache.BlockChain()
	tblock := confChain.Top() //获取创世块

	//blockLen := p.blockCache.ConfirmedLength()
	//fmt.Println(blockLen)

	//blockNum := 1000
	slot := consensus_common.GetCurrentTimestamp().Slot

	for i := 0; i < n; i++ {
		var hash []byte
		if len(blockPool) == 0 {
			//用创世块的头hash赋值
			hash = tblock.Head.Hash()
		} else {
			hash = blockPool[len(blockPool)-1].Head.Hash()
		}
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

		headInfo := generateHeadInfo(blk.Head)
		sig, _ := common.Sign(common.Secp256k1, headInfo, accountList[i%3].Seckey)
		blk.Head.Signature = sig.Encode()

		for i := 0; i < txCnt; i++ {
			blk.Content = append(blk.Content, genTx(p, i))
		}
		blockPool = append(blockPool, &blk)
	}
	return
}
func benchAddBlockCache(b *testing.B, txCnt int, continuity bool) {

	p, accountList, witnessList := envInit(b)
	//生成block
	blockPool := genBlocks(p, accountList, witnessList, b.N, txCnt, continuity)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		p.blockCache.Add(blockPool[i], p.blockVerify)
		b.StopTimer()
	}

}

// 获取block性能测试
func benchGetBlock(b *testing.B, txCnt int, continuity bool) {
	p, accountList, witnessList := envInit(b)
	//生成block
	blockPool := genBlocks(p, accountList, witnessList, b.N, txCnt, continuity)
	for i := 0; i < b.N; i++ {
		for _, bl := range blockPool {
			p.blockCache.Add(bl, p.blockVerify)
		}
	}

	//get block
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chain := p.blockCache.LongestChain()
		b.StartTimer()
		chain.GetBlockByNumber(uint64(i))
		b.StopTimer()
	}
}

// block验证性能测试
func benchBlockVerifier(b *testing.B) {
	p, accountList, witnessList := envInit(b)
	//生成block
	blockPool := genBlocks(p, accountList, witnessList, 2, 6000, true)
	//p.update(&blockPool[0].Head)
	confChain := p.blockCache.BlockChain()
	tblock := confChain.Top() //获取创世块

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.blockVerify(blockPool[0], tblock, state.StdPool)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func benchTxCache(b *testing.B, f bool) {
	p, _, _ := envInit(b)
	var txs []tx.Tx
	txCache := tx.NewTxPoolImpl()
	for j := 0; j < b.N; j++ {
		_tx := genTx(p, j)
		txs = append(txs, _tx)
	}

	b.ResetTimer()
	if f == true {
		for i := 0; i < b.N; i++ {
			b.StartTimer()
			txCache.Add(&txs[i])
			b.StopTimer()
		}
	} else {
		for i := 0; i < b.N; i++ {
			txCache.Add(&txs[i])
		}
		for i := 0; i < b.N; i++ {
			b.StartTimer()
			txCache.Del(&txs[i])
			b.StopTimer()
		}

	}
}

func benchTxCachePara(b *testing.B) {
	p, _, _ := envInit(b)
	var txs []tx.Tx

	b.ResetTimer()
	txCache := tx.NewTxPoolImpl()
	for j := 0; j < 2000; j++ {
		_tx := genTx(p, j)
		txs = append(txs, _tx)
		if j < 1000 {
			txCache.Add(&_tx)
		}
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		start := time.Now().UnixNano()
		for j := 0; j < 1000; j++ {
			txCache.Del(&txs[j])
		}
		end := time.Now().UnixNano()
		fmt.Println((end-start)/1000, " ns/op")
		wg.Done()
	}()
	go func() {
		start := time.Now().UnixNano()
		for j := 1000; j < 2000; j++ {
			txCache.Add(&txs[j])
		}
		end := time.Now().UnixNano()
		fmt.Println((end-start)/1000, " ns/op")
		wg.Done()
	}()
	wg.Wait()
}

func benchTxDb(b *testing.B, f bool) {
	p, _, _ := envInit(b)
	var txs []tx.Tx
	txDb := tx.TxDbInstance()
	for j := 0; j < b.N; j++ {
		_tx := genTx(p, j)
		txs = append(txs, _tx)
	}

	b.ResetTimer()
	if f == true {
		for i := 0; i < b.N; i++ {
			b.StartTimer()
			txDb.Add(&txs[i])
			b.StopTimer()
		}
	} else {
		for i := 0; i < b.N; i++ {
			txDb.Add(&txs[i])
		}
		for i := 0; i < b.N; i++ {
			b.StartTimer()
			txDb.Del(&txs[i])
			b.StopTimer()
		}

	}
}

// 生成block head性能测试
func benchBlockHead(b *testing.B) {
	p, _, _ := envInit(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		genBlockHead(p)
		b.StopTimer()
	}
}

// 生成块性能测试
func benchGenerateBlock(b *testing.B, txCnt int) {
	//p, _, _ := envInit(b)
	//TxPerBlk = txCnt
	//
	//for i := 0; i < TxPerBlk*b.N; i++ {
	//	_tx := genTx(p, i)
	//	p.blockCache.AddTx(&_tx)
	//}
	//
	//b.ResetTimer()
	//b.StopTimer()
	//for i := 0; i < b.N; i++ {
	//	bc := p.blockCache.LongestChain()
	//	pool := p.blockCache.LongestPool()
	//	b.StartTimer()
	//	p.genBlock(p.account, bc, pool)
	//	b.StopTimer()
	//}
}
