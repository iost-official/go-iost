package block

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"sync"

	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/core/tx"
)

var (
	blockLength    = []byte("BlockLength")    //blockLength -> length of ChainImpl
	blockStatePool = []byte("BlockStatePool") //blockStatePool -> state pool of the last block

	blockNumberPrefix = []byte("n") //blockNumberPrefix + block number -> block hash
	blockPrefix       = []byte("H") //blockHashPrefix + block hash -> block data
)

type ChainImpl struct {
	db     db.Database
	length uint64
	tx     tx.TxPool
	state  state.Pool // todo 分离这两部分
}

var chainImpl *ChainImpl

var once sync.Once

//NewBlockChain 创建一个blockChain实例,单例模式
func NewBlockChain() (chain Chain, error error) {

	once.Do(func() {
		ldb, err := db.DatabaseFactor("ldb")
		if err != nil {
			error = fmt.Errorf("failed to init db %v", err)
		}
		//defer ldb.Close()

		var length uint64
		var lenByte = make([]byte, 128)

		if ok, _ := ldb.Has(blockLength); ok {
			lenByte, err := ldb.Get(blockLength)
			if err != nil {
				error = fmt.Errorf("failed to Get blockLength")
			}

			length = binary.BigEndian.Uint64(lenByte)

		} else {
			fmt.Printf("blockLength not exist")
			length = 0
			binary.BigEndian.PutUint64(lenByte, length)

			err := ldb.Put(blockLength, lenByte)
			if err != nil {
				error = fmt.Errorf("failed to Put blockLength")
			}
		}

		tx, err := tx.NewTxPoolDb()
		if err != nil {
			error = fmt.Errorf("failed to NewTxPoolDb")
		}

		chainImpl = new(ChainImpl)
		chainImpl = &ChainImpl{db: ldb, length: length, state: nil, tx: tx}
	})

	return chainImpl, error
}

//Push 保存一个block到实例
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

	//todo:put all the tx of this block to the db
	for _, ctx := range block.Content {
		if err := b.tx.Add(&ctx); err != nil {
			return fmt.Errorf("failed to add tx %v", err)
		}
	}

	return nil
}

//Length 返回已经确定链的长度
func (b *ChainImpl) Length() uint64 {
	return b.length
}

//判断tx是否存在于db中
func (b *ChainImpl) HasTx(tx *tx.Tx) (bool, error) {
	return b.tx.Has(tx)
}

//通过hash获取tx
func (b *ChainImpl) GetTx(hash []byte) (*tx.Tx, error) {
	return b.tx.Get(hash)
}

//链长度加1
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

func (b *ChainImpl) getLengthBytes(length uint64) []byte {

	return []byte(strconv.FormatUint(length, 10))
}

//Top 返回已确定链的最后块
func (b *ChainImpl) Top() *Block {

	return b.GetBlockByNumber(b.length -1)
}

//GetBlockByNumber 通过区块编号查询块
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

	return rBlock
}

//GetBlockByHash 通过区块hash查询块
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

//暂不实现
func (b *ChainImpl) Iterator() ChainIterator {
	return nil
}
