package block

import (
	"encoding/binary"
	"fmt"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/db"
	"strconv"
	"sync"
	//"github.com/iost-official/prototype/log"
)

var (
	blockLength = []byte("BlockLength") //blockLength -> length of ChainImpl

	blockNumberPrefix = []byte("n") //blockNumberPrefix + block number -> block hash
	blockPrefix       = []byte("H") //blockHashPrefix + block hash -> block data
)

// ChainImpl 是已经确定block chain的结构体
type ChainImpl struct {
	db     db.Database
	length uint64
	tx     tx.TxPool
}

var BChain Chain
var once sync.Once

var LdbPath string

// GetInstance 创建一个blockChain实例,单例模式
func Instance() (Chain, error) {
	var err error

	once.Do(func() {

		ldb, er := db.NewLDBDatabase(LdbPath+"blockDB", 0, 0)
		if er != nil {
			err = fmt.Errorf("failed to init db %v", err)
			return
		}
		//defer ldb.Close()

		var length uint64
		var lenByte = make([]byte, 128)

		if ok, _ := ldb.Has(blockLength); ok {
			lenByte, er := ldb.Get(blockLength)
			if er != nil {
				err = fmt.Errorf("failed to Get blockLength")
				return
			}

			length = binary.BigEndian.Uint64(lenByte)

		} else {
			fmt.Printf("blockLength not exist")
			length = 0
			binary.BigEndian.PutUint64(lenByte, length)

			er := ldb.Put(blockLength, lenByte)
			if er != nil {
				err = fmt.Errorf("failed to Put blockLength")
				return
			}
		}

		txDb := tx.TxDb
		if txDb == nil {
			panic(fmt.Errorf("TxDb shouldn't be nil"))
		}
		if er != nil {
			err = fmt.Errorf("failed to NewTxPoolDb: [%v]", err)
			return
		}

		BChain = &ChainImpl{db: ldb, length: length, tx: txDb}
	})

	return BChain, err
}

// Push 保存一个block到实例
func (b *ChainImpl) Push(block *Block) error {

	hash := block.Hash()
	number := uint64(block.Head.Number)

	//存储区块hash
	err := b.db.Put(append(blockNumberPrefix, strconv.FormatUint(number, 10)...), hash)
	if err != nil {
		return fmt.Errorf("failed to Put block hash err[%v]", err)
	}

	//存储区块数据
	err = b.db.Put(append(blockPrefix, hash...), block.Encode())
	if err != nil {
		return fmt.Errorf("failed to Put block data")
	}

	err = b.lengthAdd()
	if err != nil {
		return fmt.Errorf("failed to lengthAdd %v", err)
	}
	/*
		////////////probe//////////////////
		log.Report(&log.MsgBlock{
			SubType:"confirm",
			BlockHeadHash:block.HeadHash(),
			BlockNum:block.Head.Number,
		})
		///////////////////////////////////
	*/

	//put all the tx of this block to txdb
	for _, ctx := range block.Content {
		if err := b.tx.Add(&ctx); err != nil {
			return fmt.Errorf("failed to add tx %v", err)
		}
		/*
			////////////probe//////////////////
			log.Report(&log.MsgTx{
				SubType:"confirm",
				TxHash:ctx.Hash(),
				Publisher:ctx.Publisher.Pubkey,
				Nonce:ctx.Nonce,
			})
			///////////////////////////////////
		*/
	}

	return nil
}

// Length 返回已经确定链的长度
func (b *ChainImpl) Length() uint64 {
	return b.length
}

// HasTx 判断tx是否存在于db中
func (b *ChainImpl) HasTx(tx *tx.Tx) (bool, error) {
	return b.tx.Has(tx)
}

// GetTx 通过hash获取tx
func (b *ChainImpl) GetTx(hash []byte) (*tx.Tx, error) {
	return b.tx.Get(hash)
}

// lengthAdd 链长度加1
func (b *ChainImpl) lengthAdd() error {
	b.length++

	var tmpByte = make([]byte, 128)
	binary.BigEndian.PutUint64(tmpByte, b.length)

	err := b.db.Put(blockLength, tmpByte)
	if err != nil {
		b.length--
		return fmt.Errorf("failed to Put blockLength")
	}

	return nil
}

// getLengthBytes 得到链长度的bytes类型
func (b *ChainImpl) getLengthBytes(length uint64) []byte {

	return []byte(strconv.FormatUint(length, 10))
}

// Top 返回已确定链的最后block
func (b *ChainImpl) Top() *Block {
	if b.length == 0 {
		return b.GetBlockByNumber(b.length)
	} else {
		return b.GetBlockByNumber(b.length - 1)
	}
}

// GetBlockByNumber 通过区块编号查询块
func (b *ChainImpl) GetBlockByNumber(number uint64) *Block {

	hash, err := b.db.Get(append(blockNumberPrefix, b.getLengthBytes(number)...))
	if err != nil {
		return nil
	}

	block, err := b.db.Get(append(blockPrefix, hash...))
	if err != nil {
		return nil
	}
	if len(block) == 0 {
		return nil
	}

	rBlock := new(Block)
	if err := rBlock.Decode(block); err != nil {
		return nil
	}
	//TODO:should calc sth after decode?
	return rBlock
}

// GetBlockByHash 通过区块hash查询块
func (b *ChainImpl) GetBlockByHash(blockHash []byte) *Block {

	block, err := b.db.Get(append(blockPrefix, blockHash...))
	if err != nil {
		return nil
	}
	if len(block) == 0 {
		return nil
	}

	rBlock := new(Block)
	if err := rBlock.Decode(block); err != nil {
		return nil
	}
	return rBlock
}

// Iterator 暂不实现
func (b *ChainImpl) Iterator() ChainIterator {
	return nil
}
