package transaction

import (
	"encoding/hex"
	"log"
	"github.com/iost-official/prototype/iostdb"
	"github.com/iost-official/prototype/tx/min_framework"
	"github.com/iost-official/prototype/p2p"
	"github.com/iost-official/prototype/core"
)

type Blockchain struct {
	tip []byte
	Db  *iostdb.LDBDatabase
}

type BlockchainIterator struct {
	currentHash []byte
	Db          *iostdb.LDBDatabase
}

// 改掉
func (bc *Blockchain) MineBlock(transactions []*Transaction, nn *p2p.NaiveNetwork) {
	var lastHash []byte

	lastHash, err := bc.Db.Get([]byte("l"))
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)

	/*err = bc.Db.Put(newBlock.Hash, newBlock.Serialize())
	if err != nil {
		log.Panic(err)
	}
	err = bc.Db.Put([]byte("l"), newBlock.Hash)
	if err != nil {
		log.Panic(err)
	}*/

	nn.Broadcast(core.Request{
		Time:    1,
		From:    "test1",
		To:      "test2",
		ReqType: 1,
		Body:    newBlock.Serialize(),
	})

	bc.tip = newBlock.Hash
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.Db}

	return bci
}

func (i *BlockchainIterator) Next() *Block {
	var block *Block

	encodedBlock, err := i.Db.Get(i.currentHash)
	if err != nil {
		log.Panic(err)
	}
	block = DeserializeBlock(encodedBlock)

	i.currentHash = block.PrevBlockHash

	return block
}

func dbExists(db *iostdb.LDBDatabase) bool {
	bo, err := db.Has([]byte("l"))
	if err != nil || !bo {
		return false
	}
	return true
}

// 创建一个有创世块的新链
func NewBlockchain(address string, db *iostdb.LDBDatabase) (*Blockchain, string) {
	if dbExists(db) == false {
		return nil, "No existing blockchain found. Create one first.\n"
	}
	var tip []byte
	/*db, err := iostdb.NewLDBDatabase(min_framework.DbFile, 0, 0)
	if err != nil {
		log.Panic(err)
	}*/

	tip, err := db.Get([]byte("l"))
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc, ""
}

// CreateBlockchain 创建一个新的区块链数据库
// address 用来接收挖出创世块的奖励
func CreateBlockchain(address string, db *iostdb.LDBDatabase, nn *p2p.NaiveNetwork) (*Blockchain, string) {
	if dbExists(db) {
		return nil, "Blockchain already exists.\n"
	}

	var tip []byte
	/*db, err := iostdb.NewLDBDatabase(min_framework.DbFile, 0, 0)
	if err != nil {
		log.Panic(err)
	}*/

	cbtx := NewCoinbaseTX(address, min_framework.GenesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)

	/*err := db.Put(genesis.Hash, genesis.Serialize())
	if err != nil {
		log.Panic(err)
	}
	err = db.Put([]byte("l"), genesis.Hash)
	if err != nil {
		log.Panic(err)
	}*/

	nn.Broadcast(core.Request{
		Time:    1,
		From:    "test1",
		To:      "test2",
		ReqType: 1,
		Body:    genesis.Serialize(),
	})

	tip = genesis.Hash
	bc := Blockchain{tip, db}
	return &bc, ""
}

// 暂缓修改
// FindUnspentTransactions 找到未花费输出的交易
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		// 从后往前，逆序
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// 如果交易输出被花费了
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				// 如果该交易输出可以被解锁，即可被花费
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// 针对成UTXO pool的实现
func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// UTXOPool
// FindSpendableOutputs 从 address 中找到至少 amount 的 UTXO
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}
