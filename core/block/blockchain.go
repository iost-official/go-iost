package block

import (
	"encoding/binary"

	"errors"
	"fmt"

	"strconv"

	"github.com/iost-official/Go-IOS-Protocol/db/kv"
)

type BlockChain struct {
	blockChainDB *kv.Storage
	length       int64
}

var (
	blockLength       = []byte("BlockLength")
	blockNumberPrefix = []byte("n")
	blockPrefix       = []byte("H")
)

func Int64ToByte(n int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(n))
	return b
}

func ByteToInt64(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}

func NewBlockChain(path string) (Chain, error) {
	levelDB, err := kv.NewStorage(path, kv.LevelDBStorage)
	if err != nil {
		return nil, fmt.Errorf("fail to init blockchaindb, %v", err)
	}
	var length int64 = 0
	ok, err := levelDB.Has(blockLength)
	if err != nil {
		return nil, fmt.Errorf("fail to check has(blocklength), %v", err)
	}
	if ok {
		lengthByte, err := levelDB.Get(blockLength)
		if err != nil || len(lengthByte) == 0 {
			return nil, errors.New("fail to get blocklength")
		}
		length = ByteToInt64(lengthByte)
	} else {
		lengthByte := Int64ToByte(0)
		tempErr := levelDB.Put(blockLength, lengthByte)
		if tempErr != nil {
			err = errors.New("fail to put blocklength")
		}
	}
	BC := &BlockChain{levelDB, length}
	BC.CheckLength()
	return BC, err
}

func (bc *BlockChain) Length() int64 {
	return bc.length
}

func (bc *BlockChain) Push(block *Block) error {
	err := bc.blockChainDB.BeginBatch()
	if err != nil {
		return errors.New("fail to begin batch")
	}
	hash := block.HeadHash()
	number := block.Head.Number
	bc.blockChainDB.Put(append(blockNumberPrefix, Int64ToByte(number)...), hash)
	blockByte, err := block.Encode()
	if err != nil {
		return errors.New("fail to encode block")
	}
	bc.blockChainDB.Put(append(blockPrefix, hash...), blockByte)
	bc.blockChainDB.Put(blockLength, Int64ToByte(number+1))
	err = bc.blockChainDB.CommitBatch()
	if err != nil {
		return fmt.Errorf("fail to put block, err:%s", err)
	}
	bc.length = number + 1
	return nil
}

func (bc *BlockChain) CheckLength() {
	for i := bc.length; i > 0; i-- {
		_, err := bc.GetBlockByNumber(i - 1)
		if err != nil {
			fmt.Println("fail to get the block")
		} else {
			bc.blockChainDB.Put(blockLength, Int64ToByte(i))
			bc.length = i
			break
		}
	}
}

func (bc *BlockChain) Top() (*Block, error) {
	if bc.length == 0 {
		return nil, errors.New("no block in blockChaindb")
	}
	return bc.GetBlockByNumber(bc.length - 1)
}

func (bc *BlockChain) GetHashByNumber(number int64) ([]byte, error) {
	hash, err := bc.blockChainDB.Get(append(blockNumberPrefix, Int64ToByte(number)...))
	if err != nil || len(hash) == 0 {
		return nil, errors.New("fail to get hash by number")
	}
	return hash, nil
}

func (bc *BlockChain) GetBlockByteByHash(hash []byte) ([]byte, error) {
	blockByte, err := bc.blockChainDB.Get(append(blockPrefix, hash...))
	if err != nil || len(blockByte) == 0 {
		return nil, errors.New("fail to get block byte by hash")
	}
	return blockByte, nil
}

func (bc *BlockChain) GetBlockByHash(hash []byte) (*Block, error) {
	blockByte, err := bc.GetBlockByteByHash(hash)
	if err != nil {
		return nil, err
	}
	var block Block
	err = block.Decode(blockByte)
	if err != nil {
		return nil, errors.New("fail to decode blockByte")
	}
	return &block, nil
}

func (bc *BlockChain) GetBlockByNumber(number int64) (*Block, error) {
	hash, err := bc.GetHashByNumber(number)
	if err != nil {
		return nil, err
	}
	return bc.GetBlockByHash(hash)
}

func (bc *BlockChain) Close() {
	bc.blockChainDB.Close()
}

func (bc *BlockChain) Draw(start int64, end int64) string {
	ret := ""
	for i := start; i <= end; i++ {
		blk, err := bc.GetBlockByNumber(i)
		if err != nil {
			continue
		}
		ret += strconv.FormatInt(blk.Head.Number, 10) + "(" + blk.Head.Witness[4:6] + ")-"
	}
	ret = ret[0 : len(ret)-1]
	return ret
}
