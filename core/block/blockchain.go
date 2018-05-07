package block

import (
	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/core/state"
	"fmt"
	"encoding/binary"
)

const DBPath = "levelDB/"

var (
	blockLength = []byte("BlockLength") //blockLength -> length of ChainImpl

	blockNumberPrefix = []byte("n") //blockNumberPrefix + block number -> block hash
	blockHashPrefix   = []byte("H") //blockHashPrefix + block hash -> block data
)

type ChainImpl struct {
	db     db.Database
	length uint64
	state  state.Pool
}

//NewBlockChain 创建blockChain实例
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

//Push save the block to the db
func (b *ChainImpl) Push(block *Block) error {

	return nil
}

//Length return length confirmed
func (b *ChainImpl) Length() int {
	return 0
}

//Top return the last block
func (b *ChainImpl) Top() *Block {
	return &Block{}
}

//GetBlockByNumber return the block by block number
func (b *ChainImpl) GetBlockByNumber(number int32) *Block {

	return nil
}

//GetBlockByHash return the block by block hash
func (b *ChainImpl) GetBlockByHash(blockHash []byte) *Block {
	return nil
}

//GetStatePool return confirmed state pool
func (b *ChainImpl) GetStatePool() state.Pool {
	return nil
}

//SetStatePool set confirmed state pool
func (b *ChainImpl) SetStatePool(pool state.Pool) {

}

func (b *ChainImpl) Iterator() ChainIterator {
	return nil
}
