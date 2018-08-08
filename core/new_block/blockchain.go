package block

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"sync"

	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/log"
)

var (
	blockLength = []byte("BlockLength") //blockLength -> length of ChainImpl

	blockNumberPrefix = []byte("n") //blockNumberPrefix + block number -> block hash
	blockPrefix       = []byte("H") //blockHashPrefix + block hash -> block data
)

type ChainImpl struct {
	db     *db.LDB
	length uint64
}

var BChain Chain
var once sync.Once

var LdbPath string

func Instance() (Chain, error) {
	var err error

	once.Do(func() {

		ldb, er := db.NewLDB(LdbPath+"blockDB", 0, 0)
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
		BChain = &ChainImpl{db: ldb, length: length}
		BChain.CheckLength()
	})

	return BChain, err
}

func (b *ChainImpl) Push(block *Block) error {
	btch := b.db.Batch()
	hash := block.HeadHash()
	number := uint64(block.Head.Number)
	btch.Put(append(blockNumberPrefix, strconv.FormatUint(number, 10)...), hash)
	btch.Put(append(blockPrefix, hash...), block.Encode())

	log.Log.E("[block] lengthAdd length:%v block num:%v ", b.length, number)
	l := number + 1
	var tmpByte = make([]byte, 8)
	binary.BigEndian.PutUint64(tmpByte, l)
	btch.Put(blockLength, tmpByte)

	err := btch.Commit()
	if err != nil {
		return fmt.Errorf("failed to Put block data")
	}
	b.length = l
	return nil
}

func (b *ChainImpl) Length() uint64 {
	return b.length
}

func (b *ChainImpl) CheckLength() error {
	dbLen := b.Length()
	var i uint64
	for i = dbLen; i > 0; i-- {
		_, err := b.GetBlockByNumber(i - 1)
		if err == nil {
			log.Log.I("[block] set block length %v", i)
			b.setLength(i)
			break
		} else {
			log.Log.E("[block] Length error %v", i)
		}
	}

	return nil
}

func (b *ChainImpl) setLength(l uint64) error {
	var lenB = make([]byte, 128)
	binary.BigEndian.PutUint64(lenB, l)
	er := b.db.Put(blockLength, lenB)
	if er != nil {
		return fmt.Errorf("failed to Put blockLength err:%v", er)
	}
	b.length = l
	return nil
}

func (b *ChainImpl) getLengthBytes(length uint64) []byte {
	return []byte(strconv.FormatUint(length, 10))
}

func (b *ChainImpl) Top() (*Block, error) {
	var blk *Block
	var err error
	if b.length == 0 {
		blk, err = b.GetBlockByNumber(b.length)
		if err != nil {
			return nil, err
		}
		return blk, nil
	} else {
		for i := b.length; i > 0; i-- {
			blk, err = b.GetBlockByNumber(i - 1)
			if err == nil {
				break
			}
		}
		return blk, nil
	}
}

func (b *ChainImpl) GetHashByNumber(number uint64) ([]byte, error) {
	hash, err := b.db.Get(append(blockNumberPrefix, b.getLengthBytes(number)...))
	if err != nil {
		log.Log.E("Get block hash error: %v number: %v", err, number)
		return nil, err
	}
	return hash, nil
}

func (b *ChainImpl) GetBlockByNumber(number uint64) (*Block, error) {
	hash, err := b.db.Get(append(blockNumberPrefix, b.getLengthBytes(number)...))
	if err != nil {
		log.Log.E("Get block hash error: %v number: %v", err, number)
		return nil, err
	}

	block, err := b.db.Get(append(blockPrefix, hash...))
	if err != nil {
		log.Log.E("Get block error: %v number: %v", err, number)
		return nil, err
	}
	if len(block) == 0 {
		log.Log.E("GetBlockByNumber Block empty! number: %v", number)
		return nil, fmt.Errorf("GetBlockByNumber Block empty! number: %v", number)
	}
	rBlock := new(Block)
	if err := rBlock.Decode(block); err != nil {
		log.Log.E("Failed to GetBlockByNumber Decode err: %v", err)
		return nil, err
	}
	return rBlock, nil
}

func (b *ChainImpl) GetBlockByHash(blockHash []byte) (*Block, error) {
	block, err := b.db.Get(append(blockPrefix, blockHash...))
	if err != nil {
		return nil, err
	}
	if len(block) == 0 {
		return nil, fmt.Errorf("GetBlockByHash Block empty! hash: %v", blockHash)
	}
	rBlock := new(Block)
	if err := rBlock.Decode(block); err != nil {
		return nil, err
	}
	return rBlock, nil
}

func (b *ChainImpl) GetBlockByteByHash(blockHash []byte) ([]byte, error) {
	block, err := b.db.Get(append(blockPrefix, blockHash...))
	if err != nil {
		log.Log.E("Get block error: %v hash: %v", err, string(blockHash))
		return nil, err
	}
	if len(block) == 0 {
		log.Log.E("GetBlockByteByHash Block empty! : %v", string(blockHash))
		return nil, fmt.Errorf("block empty")
	}
	return block, nil
}
