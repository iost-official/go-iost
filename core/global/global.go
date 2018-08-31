package global

import (
	"fmt"
	"time"

	"strconv"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/consensus/verifier"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/crypto"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/vm"
)

var (
	StateDBFlushTime int64 = 10000
)

func (m TMode) String() string {
	switch m {
	case ModeNormal:
		return "ModeNormal"
	case ModeSync:
		return "ModeSync"
	default:
		return ""
	}
}

type TMode uint

const (
	ModeNormal TMode = iota
	ModeSync
)

type BaseVariableImpl struct {
	blockChain block.Chain
	stateDB    db.MVCCDB
	txDB       tx.TxDB
	mode       TMode
	config     *common.Config
}

func GenGenesis(initTime int64, db db.MVCCDB) (*block.Block, error) {
	var acts []*tx.Action
	for _, k := range account.WitnessList {
		act := tx.NewAction("iost.system", "IssueIOST", fmt.Sprintf(`["%v", %v]`, k, strconv.FormatInt(account.GenesisAccount[k], 10)))
		acts = append(acts, &act)
	}
	trx := tx.NewTx(acts, nil, 0, 0, 0)
	trx.Time = 0
	acc, err := account.NewAccount(common.Base58Decode("BQd9x7rQk9Y3rVWRrvRxk7DReUJWzX4WeP9H9H4CV8Mt"))

	if err != nil {
		return nil, err
	}
	trx, err = tx.SignTx(trx, acc)
	if err != nil {
		return nil, err
	}
	blockHead := block.BlockHead{
		Version:    0,
		ParentHash: nil,
		Number:     0,
		Witness:    acc.ID,
		Time:       initTime,
	}
	engine := vm.NewEngine(&blockHead, db)
	txr, err := engine.Exec(&trx)
	if err != nil {
		return nil, fmt.Errorf("statedb push genesis failed, stop the pogram. err: %v", err)
	}
	blk := block.Block{
		Head:     &blockHead,
		Sign:     &crypto.Signature{},
		Txs:      []*tx.Tx{&trx},
		Receipts: []*tx.TxReceipt{txr},
	}
	blk.Head.TxsHash = blk.CalculateTxsHash()
	blk.Head.MerkleHash = blk.CalculateMerkleHash()
	err = blk.CalculateHeadHash()
	if err != nil {
		return nil, err
	}
	db.Tag(string(blk.HeadHash()))
	return &blk, nil
}

func New(conf *common.Config) (*BaseVariableImpl, error) {
	blockChain, err := block.NewBlockChain(conf.DB.LdbPath + "BlockChainDB")
	if err != nil {
		return nil, fmt.Errorf("new blockchain failed, stop the program. err: %v", err)
	}
	stateDB, err := db.NewMVCCDB(conf.DB.LdbPath + "StateDB")
	if err != nil {
		return nil, fmt.Errorf("new statedb failed, stop the program. err: %v", err)
	}
	blk, err := blockChain.Top()
	if err != nil {
		t := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		blk, err = GenGenesis(common.GetTimestamp(t.Unix()).Slot, stateDB)
		if err != nil {
			return nil, fmt.Errorf("new GenGenesis failed, stop the program. err: %v", err)
		}
		err = blockChain.Push(blk)
		if err != nil {
			return nil, fmt.Errorf("push block in blockChain failed, stop the program. err: %v", err)
		}
		err = stateDB.Flush(string(blk.HeadHash()))
	}
	hash := stateDB.CurrentTag()
	blk, err = blockChain.GetBlockByHash([]byte(hash))
	if err != nil {
		return nil, fmt.Errorf("get block by hash failed, stop the program. err: %v", err)
	}
	for blk.Head.Number+1 < blockChain.Length() {
		blk, err = blockChain.GetBlockByNumber(blk.Head.Number + 1)
		if err != nil {
			return nil, fmt.Errorf("get block by number failed, stop the pogram. err: %v", err)
		}
		err = verifier.VerifyBlockWithVM(blk, stateDB)
		if err != nil {
			return nil, fmt.Errorf("verify block with VM failed, stop the pogram. err: %v", err)
		}
		stateDB.Tag(string(blk.HeadHash()))
		err = stateDB.Flush(string(blk.HeadHash()))
		if err != nil {
			return nil, fmt.Errorf("flush stateDB failed, stop the pogram. err: %v", err)
		}
	}
	txDB, err := tx.NexTxDB(conf.DB.LdbPath + "TXDB")
	if err != nil {
		return nil, fmt.Errorf("new txDB failed, stop the program")
	}
	n := &BaseVariableImpl{blockChain: blockChain, stateDB: stateDB, txDB: txDB, mode: ModeNormal, config: conf}
	return n, nil
}

func FakeNew() BaseVariable {
	blockChain, _ := block.NewBlockChain("./db/BlockChainDB")
	stateDB, _ := db.NewMVCCDB("./db/StateDB")
	txDB, _ := tx.NexTxDB("./db/TXDB")
	config := common.Config{}
	return &BaseVariableImpl{blockChain, stateDB, txDB, ModeNormal, &config}
}

func (g *BaseVariableImpl) TxDB() tx.TxDB {
	return g.txDB
}

func (g *BaseVariableImpl) StateDB() db.MVCCDB {
	return g.stateDB
}

func (g *BaseVariableImpl) BlockChain() block.Chain {
	return g.blockChain
}

func (g *BaseVariableImpl) Config() *common.Config {
	return g.config
}

func (g *BaseVariableImpl) Mode() TMode {
	return g.mode
}

func (g *BaseVariableImpl) SetMode(m TMode) {
	g.mode = m
}
