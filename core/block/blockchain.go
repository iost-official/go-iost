package block

import (
	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/core/state"
	"fmt"
	"encoding/binary"
)


var (
	blockLength = []byte("BlockLength") //blockLength -> length of ChainImpl

	blockNumberPrefix = []byte("n") //blockNumberPrefix + block number -> block hash
	blockPrefix   = []byte("H") //blockHashPrefix + block hash -> block data
)

type ChainImpl struct {
	db     db.Database
	length uint64
	state  state.Pool
}

//NewBlockChain 创建一个blockChain实例
func NewBlockChain() (Chain, error) {

	ldb, err := db.DatabaseFactor("ldb")
	if err != nil {
		return nil, fmt.Errorf("failed to init db %v",err)
	}
	defer ldb.Close()

	var length uint64
	var lenByte = make([]byte, 128)

	if ok, _:=ldb.Has(blockLength); ok{
		lenByte ,err:= ldb.Get(blockLength)
		if err!=nil {
			return nil, fmt.Errorf("failed to Get blockLength")
		}

		length = binary.BigEndian.Uint64(lenByte)

	}else {
		fmt.Printf("blockLength not exist")
		length = 0
		binary.BigEndian.PutUint64(lenByte, length)

		err := ldb.Put(blockLength, lenByte)
		if err!= nil {
			return nil, fmt.Errorf("failed to Put blockLength")
		}
	}

	return &ChainImpl{db:ldb,length:length,state:nil}, nil
}

//Push 保存一个block到实例
func (b *ChainImpl) Push(block *Block) error {

	hash := block.Hash()
	number := uint64(block.Head.Number)

	var tmpByte = make([]byte, 128)
	binary.BigEndian.PutUint64(tmpByte, number)

	//存储区块hash
	err:=b.db.Put(append(blockNumberPrefix, tmpByte...),hash)
	if err!=nil {
		return fmt.Errorf("failed to Put block hash")
	}

	//存储区块数据
	err = b.db.Put(append(blockPrefix, hash...), block.Encode())
	if err!=nil {
		return fmt.Errorf("failed to Put block data")
	}

	err = b.lengthAdd()
	if err != nil {
		return fmt.Errorf("failed to lengthAdd %v", err)
	}

	//put all the transactions of this block to the ldb
	err=block.PushTxs()
	if err!=nil {
		return fmt.Errorf("%v",err)
	}
	return nil
}

//Length return length confirmed
func (b *ChainImpl) Length() uint64 {
	return b.length
}

func (b *ChainImpl) lengthAdd() error {
	b.length++

	var tmpByte = make([]byte, 128)
	binary.BigEndian.PutUint64(tmpByte, b.length)

	err := b.db.Put(blockLength, tmpByte)
	if err!= nil {
		return fmt.Errorf("failed to Put blockLength")
	}

	return nil
}

func (b *ChainImpl) getLengthBytes(length uint64) []byte {

	var tmpByte = make([]byte, 128)
	binary.BigEndian.PutUint64(tmpByte, length)

	return tmpByte
}

//Top return the last block
func (b *ChainImpl) Top() *Block {

	hash,err := b.db.Get(append(blockNumberPrefix, b.getLengthBytes(b.length)...))
	if err != nil {
		return nil
	}

	block,err := b.db.Get(append(blockPrefix, hash...))
	if err != nil {
		return nil
	}
	if len(block) == 0 {
		return nil
	}

	rBlock := new(Block)
	if err:=rBlock.Decode(block);err!=nil{
		return nil
	}

	return rBlock
}

//GetBlockByNumber return the block by block number
func (b *ChainImpl) GetBlockByNumber(number uint64) *Block {

	hash,err := b.db.Get(append(blockNumberPrefix, b.getLengthBytes(number)...))
	if err != nil {
		return nil
	}

	block,err := b.db.Get(append(blockPrefix, hash...))
	if err != nil {
		return nil
	}
	if len(block) == 0 {
		return nil
	}

	rBlock := new(Block)
	if err:=rBlock.Decode(block);err!=nil{
		return nil
	}

	return nil
}

//GetBlockByHash return the block by block hash
func (b *ChainImpl) GetBlockByHash(blockHash []byte) *Block {

	block,err := b.db.Get(append(blockPrefix, blockHash...))
	if err != nil {
		return nil
	}
	if len(block) == 0 {
		return nil
	}

	rBlock := new(Block)
	if err:=rBlock.Decode(block);err!=nil{
		return nil
	}
	return nil
}

//GetStatePool return confirmed state pool
func (b *ChainImpl) GetStatePool() state.Pool {
	return b.state
}

//SetStatePool set confirmed state pool
func (b *ChainImpl) SetStatePool(pool state.Pool) {
	b.state = pool
}

//暂不实现
func (b *ChainImpl) Iterator() ChainIterator {
	return nil
}
