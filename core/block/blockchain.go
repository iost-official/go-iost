package block

import (
	"errors"
	"fmt"
	"sync"

	"github.com/iost-official/go-iost/ilog"

	"strconv"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/db/kv"
)

// BlockChain is the implementation of chain
type BlockChain struct { //nolint:golint
	blockChainDB *kv.Storage
	rw           sync.RWMutex
	length       int64
	txTotal      int64
}

var (
	blockLength       = []byte("BlockLength")
	blockTxTotal      = []byte("BlockTxTotal")
	blockNumberPrefix = []byte("n")
	blockPrefix       = []byte("H")
	txPrefix          = []byte("t")      // txPrefix + tx hash -> block hash + tx hash
	bTxPrefix         = []byte("B")      // bTxPrefix + block hash + tx hash -> tx data
	txReceiptPrefix   = []byte("h")      // txReceiptPrefix + tx hash -> block hash + receipt hash
	receiptPrefix     = []byte("r")      // receiptPrefix + receipt hash -> block hash + receipt hash
	bReceiptPrefix    = []byte("b")      // bReceiptPrefix + block hash + receipt hash -> receipt data
	delaytxPrefix     = []byte("delay-") // delaytxPrefix + tx hash -> tx data
)

// NewBlockChain returns a Chain instance
func NewBlockChain(path string) (Chain, error) {
	levelDB, err := kv.NewStorage(path, kv.LevelDBStorage)
	if err != nil {
		return nil, fmt.Errorf("fail to init blockchaindb, %v", err)
	}
	var length int64
	var txTotal int64
	ok, err := levelDB.Has(blockLength)
	if err != nil {
		return nil, fmt.Errorf("fail to check has(blocklength), %v", err)
	}
	if ok {
		lengthByte, err := levelDB.Get(blockLength)
		if err != nil || len(lengthByte) == 0 {
			return nil, errors.New("fail to get blocklength")
		}
		length = common.BytesToInt64(lengthByte)
		txTotalByte, err := levelDB.Get(blockTxTotal)
		if err != nil || len(txTotalByte) == 0 {
			return nil, errors.New("fail to get tx total")
		}
		txTotal = common.BytesToInt64(txTotalByte)
	} else {
		lengthByte := common.Int64ToBytes(0)
		if err := levelDB.Put(blockLength, lengthByte); err != nil {
			return nil, errors.New("fail to put blocklength")
		}
		txTotalByte := common.Int64ToBytes(0)
		if err := levelDB.Put(blockTxTotal, txTotalByte); err != nil {
			return nil, errors.New("fail to put tx total")
		}
	}
	BC := &BlockChain{
		blockChainDB: levelDB,
		length:       length,
		txTotal:      txTotal,
	}
	BC.CheckLength()
	return BC, err
}

// SetLength sets blockchain's length.
func (bc *BlockChain) SetLength(l int64) {
	bc.rw.Lock()
	oldLength := bc.length
	bc.length = l
	bc.rw.Unlock()
	for i := l; i < oldLength; i++ {
		err := bc.delBlockByNumber(i)
		if err != nil {
			ilog.Error(err)
		}
	}
}

// SetTxTotal sets blockchain's tx total.
func (bc *BlockChain) SetTxTotal(i int64) {
	bc.rw.Lock()
	bc.txTotal = i
	bc.rw.Unlock()
}

// Length return length of block chain
func (bc *BlockChain) Length() int64 {
	bc.rw.RLock()
	defer bc.rw.RUnlock()
	return bc.length
}

// TxTotal return tx total of block chain
func (bc *BlockChain) TxTotal() int64 {
	bc.rw.RLock()
	defer bc.rw.RUnlock()
	return bc.txTotal
}

// Push save the block to database
func (bc *BlockChain) Push(block *Block) error {
	err := bc.blockChainDB.BeginBatch()
	if err != nil {
		return errors.New("fail to begin batch")
	}

	hash := block.HeadHash()
	number := block.Head.Number
	txTotal := bc.TxTotal()
	bc.blockChainDB.Put(append(blockNumberPrefix, common.Int64ToBytes(number)...), hash)
	blockByte, err := block.EncodeM()
	if err != nil {
		return errors.New("fail to encode block")
	}
	bc.blockChainDB.Put(append(blockPrefix, hash...), blockByte)
	bc.blockChainDB.Put(blockLength, common.Int64ToBytes(number+1))
	bc.blockChainDB.Put(blockTxTotal, common.Int64ToBytes(txTotal+int64(len(block.Txs))))
	for i, t := range block.Txs {
		tHash := t.Hash()
		txBytes := t.Encode()
		bc.blockChainDB.Put(append(txPrefix, tHash...), append(hash, tHash...))
		bc.blockChainDB.Put(append(bTxPrefix, append(hash, tHash...)...), txBytes)

		// save receipt
		rHash := block.Receipts[i].Hash()
		bc.blockChainDB.Put(append(txReceiptPrefix, tHash...), append(hash, rHash...))
		bc.blockChainDB.Put(append(receiptPrefix, rHash...), append(hash, rHash...))
		bc.blockChainDB.Put(append(bReceiptPrefix, append(hash, rHash...)...), block.Receipts[i].Encode())

		if t.Delay > 0 && block.Receipts[i].Status.Code == tx.Success {
			bc.blockChainDB.Put(append(delaytxPrefix, tHash...), txBytes)
		}
		if t.IsDefer() {
			bc.blockChainDB.Delete(append(delaytxPrefix, t.ReferredTx...))
		}

		canceledDelayHashes := block.Receipts[i].ParseCancelDelaytx()
		for _, canceledHash := range canceledDelayHashes {
			bc.blockChainDB.Delete(append(delaytxPrefix, canceledHash...))
		}
	}
	err = bc.blockChainDB.CommitBatch()
	if err != nil {
		return fmt.Errorf("fail to put block, err:%s", err)
	}
	bc.SetLength(number + 1)
	bc.SetTxTotal(txTotal + int64(len(block.Txs)))
	return nil
}

// CheckLength is check length of block in database
func (bc *BlockChain) CheckLength() {
	for i := bc.Length(); i > 0; i-- {
		_, err := bc.GetBlockByNumber(i - 1)
		if err != nil {
			fmt.Println("fail to get the block")
		} else {
			bc.blockChainDB.Put(blockLength, common.Int64ToBytes(i))
			bc.SetLength(i)
			break
		}
	}
}

// Top return the block
func (bc *BlockChain) Top() (*Block, error) {
	if bc.Length() == 0 {
		return nil, errors.New("no block in blockChaindb")
	}
	return bc.GetBlockByNumber(bc.Length() - 1)
}

// GetHashByNumber is get hash by number
func (bc *BlockChain) GetHashByNumber(number int64) ([]byte, error) {
	hash, err := bc.blockChainDB.Get(append(blockNumberPrefix, common.Int64ToBytes(number)...))
	if err != nil || len(hash) == 0 {
		return nil, errors.New("fail to get hash by number")
	}
	return hash, nil
}

// getBlockByteByHash is get block byte by hash
func (bc *BlockChain) getBlockByteByHash(hash []byte) ([]byte, error) {
	blockByte, err := bc.blockChainDB.Get(append(blockPrefix, hash...))
	if err != nil || len(blockByte) == 0 {
		return nil, errors.New("fail to get block byte by hash")
	}
	return blockByte, nil
}
func (bc *BlockChain) delBlockByHash(hash []byte) error {
	blockByte, err := bc.getBlockByteByHash(hash)
	if err != nil {
		return err
	}
	var blk Block
	err = blk.Decode(blockByte)
	if err != nil {
		return errors.New("fail to decode blockByte")
	}
	for _, txHash := range blk.TxHashes {
		err = bc.blockChainDB.Delete(append(txPrefix, txHash...))
		if err != nil {
			return err
		}
	}

	for _, rHash := range blk.ReceiptHashes {
		err = bc.blockChainDB.Delete(append(receiptPrefix, rHash...))
		if err != nil {
			return err
		}
	}

	err = bc.blockChainDB.Delete(append(blockPrefix, hash...))
	if err != nil {
		return err
	}
	return nil
}

// GetBlockByHash is get block by hash
func (bc *BlockChain) GetBlockByHash(hash []byte) (*Block, error) {
	blockByte, err := bc.getBlockByteByHash(hash)
	if err != nil {
		return nil, err
	}
	var blk Block
	err = blk.Decode(blockByte)
	if err != nil {
		return nil, errors.New("fail to decode blockByte")
	}
	if blk.TxHashes != nil {
		blk.Txs = make([]*tx.Tx, len(blk.TxHashes))
		txsMap, err := bc.getBlockTxsMap(hash)
		if err != nil {
			return nil, err
		}
		for i, hash := range blk.TxHashes {
			if tx, ok := txsMap[string(hash)]; ok {
				blk.Txs[i] = tx
			} else {
				return nil, fmt.Errorf("miss the tx, tx hash: %s", hash)
			}
		}
	}
	if blk.ReceiptHashes != nil {
		blk.Receipts = make([]*tx.TxReceipt, len(blk.ReceiptHashes))
		receiptMap, err := bc.getBlockReceiptMap(hash)
		if err != nil {
			return nil, err
		}
		for i, hash := range blk.ReceiptHashes {
			if tr, ok := receiptMap[string(hash)]; ok {
				blk.Receipts[i] = tr
			} else {
				return nil, fmt.Errorf("miss the tx receipt, tx receipt hash: %s", hash)
			}
		}
	}
	return &blk, nil
}

func (bc *BlockChain) delBlockByNumber(number int64) error {
	hash, err := bc.GetHashByNumber(number)
	if err != nil {
		return err
	}
	return bc.delBlockByHash(hash)
}

// GetBlockByNumber is get block by number
func (bc *BlockChain) GetBlockByNumber(number int64) (*Block, error) {
	hash, err := bc.GetHashByNumber(number)
	if err != nil {
		return nil, err
	}
	return bc.GetBlockByHash(hash)
}

func (bc *BlockChain) getBlockTxsMap(hash []byte) (map[string]*tx.Tx, error) {
	iter := bc.blockChainDB.NewIteratorByPrefix(append(bTxPrefix, hash...))
	txsMap := make(map[string]*tx.Tx, 0)
	for iter.Next() {
		var tt tx.Tx
		err := tt.Decode(iter.Value())
		if err != nil {
			return nil, fmt.Errorf("fail to decode tx: %v", err)
		}
		txsMap[string(tt.Hash())] = &tt
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return nil, fmt.Errorf("fail to get block txs: %v", err)
	}

	return txsMap, nil
}

func (bc *BlockChain) getBlockReceiptMap(hash []byte) (map[string]*tx.TxReceipt, error) {
	iter := bc.blockChainDB.NewIteratorByPrefix(append(bReceiptPrefix, hash...))
	receiptMap := make(map[string]*tx.TxReceipt, 0)
	for iter.Next() {
		var tr tx.TxReceipt
		err := tr.Decode(iter.Value())
		if err != nil {
			return nil, fmt.Errorf("fail to decode tx receipt: %v", err)
		}
		receiptMap[string(tr.Hash())] = &tr
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return nil, fmt.Errorf("fail to get block receipts: %v", err)
	}

	return receiptMap, nil
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

// Size returns the blockchain db size
func (bc *BlockChain) Size() (int64, error) {
	return bc.blockChainDB.Size()
}

// Close is close database
func (bc *BlockChain) Close() {
	bc.blockChainDB.Close()
}

// AllDelaytx returns all delay transactions.
func (bc *BlockChain) AllDelaytx() ([]*tx.Tx, error) {
	iter := bc.blockChainDB.NewIteratorByPrefix(delaytxPrefix)
	ret := make([]*tx.Tx, 0)
	for iter.Next() {
		t := &tx.Tx{}
		err := t.Decode(iter.Value())
		if err != nil {
			continue
		}
		ret = append(ret, t)
	}
	return ret, nil
}

// Draw the graph about blockchain
func (bc *BlockChain) Draw(start int64, end int64) string {
	ret := ""
	for i := start; i <= end; i++ {
		blk, err := bc.GetBlockByNumber(i)
		if err != nil {
			continue
		}
		ret += strconv.FormatInt(blk.Head.Number, 10) + "(" + blk.Head.Witness[4:6] + ")(" + strconv.Itoa(len(blk.Txs)) + ")-"
	}
	ret = ret[0 : len(ret)-1]
	return ret
}
