package global

import (
	"fmt"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/consensus/verifier"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/crypto"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/vm"
)

type TMode uint

const (
	ModeNormal TMode = iota
	ModeSync
	ModeFetchGenesis
	ModeInit
)

func (m TMode) String() string {
	switch m {
	case ModeNormal:
		return "ModeNormal"
	case ModeSync:
		return "ModeSync"
	case ModeFetchGenesis:
		return "ModeFetchGenesis"
	case ModeInit:
		return "ModeInit"
	default:
		return ""
	}
}

type BaseVariableImpl struct {
	blockChain  block.Chain
	stateDB     db.MVCCDB
	txDB        tx.TxDB
	mode        TMode
	witnessList []string
	config      *common.Config
}

func GenGenesis(db db.MVCCDB, witnessInfo []string) (*block.Block, error) {
	var acts []*tx.Action
	for i := 0; i < len(witnessInfo)/2; i++ {
		act := tx.NewAction("iost.system", "IssueIOST", fmt.Sprintf(`["%v", %v]`, witnessInfo[2*i], witnessInfo[2*i+1]))
		acts = append(acts, &act)
	}
	trx := tx.NewTx(acts, nil, 0, 0, 0)
	trx.Time = 0
	acc, err := account.NewAccount(common.Base58Decode("2vj2Ab8Taz1TT2MSQHxmSffGnvsc9EVrmjx1W7SBQthCpuykhbRn2it8DgNkcm4T9tdBgsue3uBiAzxLpLJoDUbc"), crypto.Ed25519)
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
		Time:       0,
	}
	engine := vm.NewEngine(&blockHead, db)
	txr, err := engine.Exec(trx)
	if err != nil {
		return nil, fmt.Errorf("exec tx failed, stop the pogram. err: %v", err)
	}
	blk := block.Block{
		Head:     &blockHead,
		Sign:     &crypto.Signature{},
		Txs:      []*tx.Tx{trx},
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
	var blockChain block.Chain
	var stateDB db.MVCCDB
	var txDB tx.TxDB
	var err error
	var witnessList []string
	for i := 0; i < len(conf.Genesis.WitnessInfo)/2; i++ {
		witnessList = append(witnessList, conf.Genesis.WitnessInfo[2*i])
	}
	blockChain, err = block.NewBlockChain(conf.DB.LdbPath + "BlockChainDB")
	if err != nil {
		return nil, fmt.Errorf("new blockchain failed, stop the program. err: %v", err)
	}
	blk, err := blockChain.GetBlockByNumber(0)
	if err != nil { //blockchaindb is empty
		stateDB, err = db.NewMVCCDB(conf.DB.LdbPath + "StateDB")
		if err != nil {
			return nil, fmt.Errorf("new statedb failed, stop the program. err: %v", err)
		}
		hash := stateDB.CurrentTag()
		if hash != "" {
			return nil, fmt.Errorf("blockchaindb is empty, but statedb is not")
		}
		txDB, err = tx.NewTxDB(conf.DB.LdbPath + "TXDB")
		if err != nil {
			return nil, fmt.Errorf("new txDB failed, stop the program. err: %v", err)
		}
		if conf.Genesis.CreateGenesis {
			blk, err = GenGenesis(stateDB, conf.Genesis.WitnessInfo)
			if err != nil {
				return nil, fmt.Errorf("new GenGenesis failed, stop the program. err: %v", err)
			}
			err = blockChain.Push(blk)
			if err != nil {
				return nil, fmt.Errorf("push block in blockChain failed, stop the program. err: %v", err)
			}
			err = stateDB.Flush(string(blk.HeadHash()))
			if err != nil {
				return nil, fmt.Errorf("flush block into stateDB failed, stop the program. err: %v", err)
			}
			err = txDB.Push(blk.Txs, blk.Receipts)
			if err != nil {
				return nil, fmt.Errorf("push txDB failed, stop the pogram. err: %v", err)
			}
		}
		return &BaseVariableImpl{blockChain: blockChain, stateDB: stateDB, txDB: txDB, mode: ModeInit, witnessList: witnessList, config: conf}, nil
	}
	if common.Base58Encode(blk.HeadHash()) != conf.Genesis.GenesisHash { //get data from seedNode
		return nil, fmt.Errorf("the hash of genesis block in db doesn't match that in config")
	}
	stateDB, err = db.NewMVCCDB(conf.DB.LdbPath + "StateDB")
	if err != nil {
		return nil, fmt.Errorf("new statedb failed, stop the program. err: %v", err)
	}
	hash := stateDB.CurrentTag()
	blk, err = blockChain.GetBlockByHash([]byte(hash))
	if err != nil {
		return nil, fmt.Errorf("statedb doesn't coincides with blockchaindb. err: %v", err)
	}
	for i := blk.Head.Number + 1; i < blockChain.Length(); i++ {
		blk, err = blockChain.GetBlockByNumber(i)
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
	txDB, err = tx.NewTxDB(conf.DB.LdbPath + "TXDB")
	if err != nil {
		return nil, fmt.Errorf("new txDB failed, stop the program. err: %v", err)
	}
	return &BaseVariableImpl{blockChain: blockChain, stateDB: stateDB, txDB: txDB, mode: ModeInit, witnessList: witnessList, config: conf}, nil
}

func FakeNew() BaseVariable {
	blockChain, _ := block.NewBlockChain("./db/BlockChainDB")
	stateDB, _ := db.NewMVCCDB("./db/StateDB")
	txDB, _ := tx.NewTxDB("./db/TXDB")
	config := common.Config{}
	return &BaseVariableImpl{blockChain, stateDB, txDB, ModeNormal, []string{""}, &config}
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

func (g *BaseVariableImpl) WitnessList() []string {
	return g.witnessList
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
