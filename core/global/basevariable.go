package global

import (
	"fmt"

	"os"

	"time"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/verifier"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm"
	"github.com/iost-official/go-iost/vm/native"
)

// TMode type of mode
type TMode uint

const (
	// ModeNormal is normal mode
	ModeNormal TMode = iota
	// ModeSync is sync mode
	ModeSync
	// ModeInit init mode
	ModeInit
)

// VoteContractPath is config of vote
var VoteContractPath = "../../config/"
var adminID = ""

// GenesisTxExecTime is the maximum execution time of a transaction in genesis block
var GenesisTxExecTime = 1 * time.Second

// String return string of mode
func (m TMode) String() string {
	switch m {
	case ModeNormal:
		return "ModeNormal"
	case ModeSync:
		return "ModeSync"
	case ModeInit:
		return "ModeInit"
	default:
		return ""
	}
}

// BaseVariableImpl is the implementation of BaseVariable
type BaseVariableImpl struct {
	blockChain  block.Chain
	stateDB     db.MVCCDB
	txDB        TxDB
	mode        TMode
	witnessList []string
	config      *common.Config
}

// GenGenesis is create a genesis block
func GenGenesis(db db.MVCCDB, witnessInfo []string, t common.Timestamp) (*block.Block, error) {
	var acts []*tx.Action
	for i := 0; i < len(witnessInfo)/2; i++ {
		act := tx.NewAction("iost.system", "IssueIOST", fmt.Sprintf(`["%v", %v]`, witnessInfo[2*i], witnessInfo[2*i+1]))
		acts = append(acts, &act)
	}
	// deploy iost.vote
	voteFilePath := VoteContractPath + "vote.js"
	voteAbiPath := VoteContractPath + "vote.js.abi"
	fd, err := common.ReadFile(voteFilePath)
	if err != nil {
		return nil, err
	}
	rawCode := string(fd)
	fd, err = common.ReadFile(voteAbiPath)
	if err != nil {
		return nil, err
	}
	rawAbi := string(fd)
	c := contract.Compiler{}
	code, err := c.Parse("iost.vote", rawCode, rawAbi)
	if err != nil {
		return nil, err
	}

	act := tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.vote", code.B64Encode()))
	acts = append(acts, &act)

	num := len(witnessInfo) / 2
	for i := 0; i < num; i++ {
		act1 := tx.NewAction("iost.vote", "InitProducer", fmt.Sprintf(`["%v"]`, witnessInfo[2*i]))
		acts = append(acts, &act1)
	}
	act11 := tx.NewAction("iost.vote", "InitAdmin", fmt.Sprintf(`["%v"]`, adminID))
	acts = append(acts, &act11)

	// deploy iost.bonus
	act2 := tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.bonus", native.BonusABI().B64Encode()))
	acts = append(acts, &act2)

	trx := tx.NewTx(acts, nil, 100000000, 0, 0)
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
		Time:       t.Slot,
	}
	engine := vm.NewEngine(&blockHead, db)
	txr, err := engine.Exec(trx, GenesisTxExecTime)
	if err != nil || txr.Status.Code != tx.Success {
		return nil, fmt.Errorf("exec tx failed, stop the pogram. err: %v, receipt: %v", err, txr)
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

// New return a BaseVariable instance
// nolint: gocyclo
func New(conf *common.Config) (*BaseVariableImpl, error) {
	var blockChain block.Chain
	var stateDB db.MVCCDB
	var txDB TxDB
	var err error
	var witnessList []string

	v := common.LoadYamlAsViper(conf.GenesisConfigPath)
	genesisConfig := &common.GenesisConfig{}
	if err := v.Unmarshal(genesisConfig); err != nil {
		ilog.Fatalf("Unable to decode into struct, %v", err)
	}

	VoteContractPath = genesisConfig.VoteContractPath
	adminID = genesisConfig.AdminID

	for i := 0; i < len(genesisConfig.WitnessInfo)/2; i++ {
		witnessList = append(witnessList, genesisConfig.WitnessInfo[2*i])
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
		txDB, err = NewTxDB(conf.DB.LdbPath + "TXDB")
		if err != nil {
			return nil, fmt.Errorf("new txDB failed, stop the program. err: %v", err)
		}
		if genesisConfig.CreateGenesis {
			t, err := common.ParseStringToTimestamp(genesisConfig.InitialTimestamp)
			if err != nil {
				ilog.Fatalf("invalid genesis initial time string %v (%v).", genesisConfig.InitialTimestamp, err)
			}
			blk, err = GenGenesis(stateDB, genesisConfig.WitnessInfo, t)
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
			genesisBlock, _ := blockChain.GetBlockByNumber(0)
			ilog.Infof("createGenesisHash: %v", common.Base58Encode(genesisBlock.HeadHash()))
		}
		return &BaseVariableImpl{blockChain: blockChain, stateDB: stateDB, txDB: txDB, mode: ModeInit, witnessList: witnessList, config: conf}, nil
	}
	stateDB, err = db.NewMVCCDB(conf.DB.LdbPath + "StateDB")
	if err != nil {
		return nil, fmt.Errorf("new statedb failed, stop the program. err: %v", err)
	}
	hash := stateDB.CurrentTag()
	blk, err = blockChain.GetBlockByHash([]byte(hash))
	if err != nil && hash != "" {
		return nil, fmt.Errorf("statedb doesn't coincides with blockchaindb. err: %v", err)
	}
	var startNumebr int64
	if err == nil {
		startNumebr = blk.Head.Number + 1
	}
	for i := startNumebr; i < blockChain.Length(); i++ {
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
	txDB, err = NewTxDB(conf.DB.LdbPath + "TXDB")
	if err != nil {
		return nil, fmt.Errorf("new txDB failed, stop the program. err: %v", err)
	}

	return &BaseVariableImpl{blockChain: blockChain, stateDB: stateDB, txDB: txDB, mode: ModeInit, witnessList: witnessList, config: conf}, nil
}

// FakeNew is fake BaseVariable
func FakeNew() (*BaseVariableImpl, error) {
	blockChain, err := block.NewBlockChain("./Fakedb/BlockChainDB")
	if err != nil {
		return nil, err
	}
	stateDB, err := db.NewMVCCDB("./Fakedb/StateDB")
	if err != nil {
		return nil, err
	}
	txDB, err := NewTxDB("./Fakedb/TXDB")
	if err != nil {
		return nil, err
	}
	config := common.Config{}
	config.VM = &common.VMConfig{}
	config.VM.JsPath = os.Getenv("GOPATH") + "/src/github.com/iost-official/go-iost/vm/v8vm/v8/libjs/"

	vm.SetUp(config.VM)
	VoteContractPath = os.Getenv("GOPATH") + "/src/github.com/iost-official/go-iost/config/"
	fmt.Println(VoteContractPath)
	fmt.Println(config.VM.JsPath)
	blk, err := GenGenesis(stateDB, []string{"a1", "11111111111", "a2", "2222", "a3", "333"}, common.Timestamp{})
	if err != nil {
		return nil, err
	}
	blk.CalculateHeadHash()
	blk.CalculateTxsHash()
	blk.CalculateMerkleHash()
	err = blockChain.Push(blk)
	if err != nil {
		return nil, err
	}
	err = stateDB.Flush(string(blk.HeadHash()))
	if err != nil {
		return nil, err
	}
	err = txDB.Push(blk.Txs, blk.Receipts)
	if err != nil {
		return nil, err
	}

	return &BaseVariableImpl{blockChain, stateDB, txDB, ModeNormal, []string{""}, &config}, nil
}

// TxDB return the transaction database
func (g *BaseVariableImpl) TxDB() TxDB {
	return g.txDB
}

// StateDB return the state database
func (g *BaseVariableImpl) StateDB() db.MVCCDB {
	return g.stateDB
}

// BlockChain return the block chain
func (g *BaseVariableImpl) BlockChain() block.Chain {
	return g.blockChain
}

// WitnessList return the witness list
func (g *BaseVariableImpl) WitnessList() []string {
	return g.witnessList
}

// Config return the config
func (g *BaseVariableImpl) Config() *common.Config {
	return g.config
}

// Mode return the mode
func (g *BaseVariableImpl) Mode() TMode {
	return g.mode
}

// SetMode is set the mode
func (g *BaseVariableImpl) SetMode(m TMode) {
	g.mode = m
}
