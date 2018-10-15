package block

import (
	"encoding/binary"

	"errors"
	"fmt"

	"strconv"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/db/kv"
)

// BlockChain is the implementation of chain
type BlockChain struct { //nolint:golint
	blockChainDB *kv.Storage
	length       int64
}

var (
	blockLength       = []byte("BlockLength")
	blockNumberPrefix = []byte("n")
	blockPrefix       = []byte("H")
	blockMPrefix      = []byte("M")
	txPrefix          = []byte("t") // txPrefix+tx hash -> tx data
	bTxPrefix         = []byte("B") // txPrefix+tx hash -> tx data
	txReceiptPrefix   = []byte("h") // receiptHashPrefix + tx hash -> receipt hash
	receiptPrefix     = []byte("r") // receiptPrefix + receipt hash -> receipt data
	bReceiptPrefix    = []byte("b") // txPrefix+tx hash -> tx data
)

// Int64ToByte is int64 to byte
func Int64ToByte(n int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(n))
	return b
}

// ByteToInt64 is byte to int64
func ByteToInt64(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}

// NewBlockChain returns a Chain instance
func NewBlockChain(path string) (Chain, error) {
	levelDB, err := kv.NewStorage(path, kv.LevelDBStorage)
	if err != nil {
		return nil, fmt.Errorf("fail to init blockchaindb, %v", err)
	}
	var length int64
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

// Length return length of block chain
func (bc *BlockChain) Length() int64 {
	return bc.length
}

// Push save the block to database
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
	// blockM
	blockByte, err = block.EncodeM()
	if err != nil {
		return errors.New("fail to encode block")
	}
	bc.blockChainDB.Put(append(blockMPrefix, hash...), blockByte)

	bc.blockChainDB.Put(blockLength, Int64ToByte(number+1))
	for _, tx := range block.Txs {
		tHash := tx.Hash()
		bc.blockChainDB.Put(append(txPrefix, tHash...), append(hash, tHash...))
		bc.blockChainDB.Put(append(bTxPrefix, append(hash, tHash...)...), tx.Encode())

		/*
			// save receipt
			rHash := block.Receipts[i].Hash()
			bc.blockChainDB.Put(append(txReceiptPrefix, tHash...), append(hash, rHash...))
			bc.blockChainDB.Put(append(receiptPrefix, rHash...), append(hash, rHash...))
			bc.blockChainDB.Put(append(bReceiptPrefix, append(hash, rHash...)...), receipts[i].Encode())
		*/
	}
	err = bc.blockChainDB.CommitBatch()
	if err != nil {
		return fmt.Errorf("fail to put block, err:%s", err)
	}
	bc.length = number + 1
	return nil
}

// CheckLength is check length of block in database
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

// Top return the block
func (bc *BlockChain) Top() (*Block, error) {
	if bc.length == 0 {
		return nil, errors.New("no block in blockChaindb")
	}
	return bc.GetBlockByNumber(bc.length - 1)
}

// GetHashByNumber is get hash by number
func (bc *BlockChain) GetHashByNumber(number int64) ([]byte, error) {
	hash, err := bc.blockChainDB.Get(append(blockNumberPrefix, Int64ToByte(number)...))
	if err != nil || len(hash) == 0 {
		return nil, errors.New("fail to get hash by number")
	}
	return hash, nil
}

// GetBlockByteByHash is get block byte by hash
func (bc *BlockChain) GetBlockByteByHash(hash []byte) ([]byte, error) {
	blockByte, err := bc.blockChainDB.Get(append(blockPrefix, hash...))
	if err != nil || len(blockByte) == 0 {
		return nil, errors.New("fail to get block byte by hash")
	}
	return blockByte, nil
}

// GetBlockByHash is get block by hash
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

// GetBlockByNumber is get block by number
func (bc *BlockChain) GetBlockByNumber(number int64) (*Block, error) {
	hash, err := bc.GetHashByNumber(number)
	if err != nil {
		return nil, err
	}
	return bc.GetBlockByHash(hash)
}

func (bc *BlockChain) GetBlockMByHash(hash []byte) (*Block, error) {
	blockByte, err := bc.blockChainDB.Get(append(blockMPrefix, hash...))
	if err != nil || len(blockByte) == 0 {
		return nil, errors.New("fail to get block byte by hash")
	}
	if err != nil {
		return nil, err
	}
	var block Block
	err = block.DecodeM(blockByte)
	if err != nil {
		return nil, errors.New("fail to decode blockByte")
	}
	return &block, nil
}
func (bc *BlockChain) GetBlockTxsMap(hash []byte) (map[string]*tx.Tx, error) {
	iter, err := bc.blockChainDB.Range(append(bTxPrefix, hash...))
	if err != nil {
		return nil, errors.New("fail to get block txs")
	}
	txsMap := make(map[string]*tx.Tx, 0)
	for iter.Next() {
		var tt tx.Tx
		err = tt.Decode(iter.Value())
		txsMap[string(tt.Hash())] = &tt
	}
	return txsMap, nil
}

// GetTx gets tx with tx's hash.
func (bc *BlockChain) GetTx(hash []byte) (*tx.Tx, error) {
	tx := tx.Tx{}
	bTx, err := bc.blockChainDB.Get(append(txPrefix, hash...))
	if err != nil {
		return nil, fmt.Errorf("failed to Get the tx: %v", err)
	}
	txData, err := bc.blockChainDB.Get(append(bTxPrefix, bTx...))
	if err != nil {
		return nil, fmt.Errorf("failed to Get the tx: %v", err)
	}
	if len(txData) == 0 {
		return nil, fmt.Errorf("failed to Get the tx: not found")
	}
	err = tx.Decode(txData)
	if err != nil {
		return nil, fmt.Errorf("failed to Decode the tx: %v", err)
	}
	return &tx, nil
}

// HasTx checks if database has tx.
func (bc *BlockChain) HasTx(hash []byte) (bool, error) {
	return bc.blockChainDB.Has(append(txPrefix, hash...))
}

// GetReceipt gets receipt with receipt's hash
func (bc *BlockChain) GetReceipt(hash []byte) (*tx.TxReceipt, error) {
	bReHash, err := bc.blockChainDB.Get(append(receiptPrefix, hash...))
	if err != nil {
		return nil, fmt.Errorf("failed to Get the receipt: %v", err)
	}
	reData, err := bc.blockChainDB.Get(append(bReceiptPrefix, bReHash...))
	if err != nil {
		return nil, fmt.Errorf("failed to Get the receipt: %v", err)
	}
	if len(reData) == 0 {
		return nil, fmt.Errorf("failed to Get the receipt: not found")
	}
	re := tx.TxReceipt{}
	err = re.Decode(reData)
	if err != nil {
		return nil, fmt.Errorf("failed to Decode the receipt: %v", err)
	}
	return &re, nil
}

// GetReceiptByTxHash gets receipt with tx's hash
func (bc *BlockChain) GetReceiptByTxHash(hash []byte) (*tx.TxReceipt, error) {
	bReHash, err := bc.blockChainDB.Get(append(txReceiptPrefix, hash...))
	if err != nil {
		return nil, fmt.Errorf("failed to Get the receipt hash: %v", err)
	}
	reData, err := bc.blockChainDB.Get(append(bReceiptPrefix, bReHash...))
	if err != nil {
		return nil, fmt.Errorf("failed to Get the receipt: %v", err)
	}
	if len(reData) == 0 {
		return nil, fmt.Errorf("failed to Get the receipt: not found")
	}
	re := tx.TxReceipt{}
	err = re.Decode(reData)
	if err != nil {
		return nil, fmt.Errorf("failed to Decode the receipt: %v", err)
	}
	return &re, nil
}

// HasReceipt checks if database has receipt.
func (bc *BlockChain) HasReceipt(hash []byte) (bool, error) {
	return bc.blockChainDB.Has(append(receiptPrefix, hash...))
}

// Close is close database
func (bc *BlockChain) Close() {
	bc.blockChainDB.Close()
}

// Draw the graph about blockchain
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
