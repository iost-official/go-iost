package genesis

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/verifier"
	"github.com/iost-official/go-iost/vm"
	"github.com/iost-official/go-iost/vm/native"
)

// GenesisTxExecTime is the maximum execution time of a transaction in genesis block
var GenesisTxExecTime = 1 * time.Second

// GenGenesisByFile is create a genesis block by config file
func GenGenesisByFile(db db.MVCCDB, path string) (*block.Block, error) {
	v := common.LoadYamlAsViper(path)
	genesisConfig := &common.GenesisConfig{}
	if err := v.Unmarshal(genesisConfig); err != nil {
		ilog.Fatalf("Unable to decode into struct, %v", err)
	}
	return GenGenesis(db, genesisConfig)
}

// GenGenesis is create a genesis block
func GenGenesis(db db.MVCCDB, gConf *common.GenesisConfig) (*block.Block, error) {
	witnessInfo := gConf.WitnessInfo
	t, err := common.ParseStringToTimestamp(gConf.InitialTimestamp)
	if err != nil {
		ilog.Fatalf("invalid genesis initial time string %v (%v).", gConf.InitialTimestamp, err)
	}
	var acts []*tx.Action
	for _, v := range witnessInfo {
		act := tx.NewAction("iost.system", "IssueIOST", fmt.Sprintf(`["%v", %v]`, v.ID, v.Balance))
		acts = append(acts, &act)
	}
	// deploy iost.vote
	voteFilePath := filepath.Join(gConf.VoteContractPath, "vote.js")
	voteAbiPath := filepath.Join(gConf.VoteContractPath, "vote.js.abi")
	fd, err := ioutil.ReadFile(voteFilePath)
	if err != nil {
		return nil, err
	}
	rawCode := string(fd)
	fd, err = ioutil.ReadFile(voteAbiPath)
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

	for _, v := range witnessInfo {
		act1 := tx.NewAction("iost.vote", "InitProducer", fmt.Sprintf(`["%v"]`, v.Owner))
		acts = append(acts, &act1)
	}
	act11 := tx.NewAction("iost.vote", "InitAdmin", fmt.Sprintf(`["%v"]`, gConf.AdminID))
	acts = append(acts, &act11)

	// deploy iost.bonus
	act2 := tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.bonus", native.BonusABI().B64Encode()))
	acts = append(acts, &act2)
	// deploy iost.gas
	act3 := tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.gas", native.GasABI().B64Encode()))
	acts = append(acts, &act3)

	trx := tx.NewTx(acts, nil, 100000000, 0, 0)
	trx.Time = 0
	acc, err := account.NewKeyPair(common.Base58Decode("2vj2Ab8Taz1TT2MSQHxmSffGnvsc9EVrmjx1W7SBQthCpuykhbRn2it8DgNkcm4T9tdBgsue3uBiAzxLpLJoDUbc"), crypto.Ed25519) // TODO 修改为account
	if err != nil {
		return nil, err
	}
	trx, err = tx.SignTx(trx, acc.ID, acc)
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
	v := verifier.Verifier{}
	txr, err := v.Exec(&blockHead, db, trx, GenesisTxExecTime)
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

// FakeBv is fake BaseVariable
func FakeBv(bv global.BaseVariable) error {
	config := common.Config{}
	config.VM = &common.VMConfig{}
	config.VM.JsPath = os.Getenv("GOPATH") + "/src/github.com/iost-official/go-iost/vm/v8vm/v8/libjs/"

	vm.SetUp(config.VM)

	blk, err := GenGenesis(
		bv.StateDB(),
		&common.GenesisConfig{
			WitnessInfo: []*common.Witness{
				{ID: "a1", Owner: "a1", Active: "a1", Balance: 11111111111},
				{ID: "a2", Owner: "a2", Active: "a2", Balance: 222222},
				{ID: "a3", Owner: "a3", Active: "a3", Balance: 333333333}},
			InitialTimestamp: "2006-01-02T15:04:05Z",
			VoteContractPath: os.Getenv("GOPATH") + "/src/github.com/iost-official/go-iost/config/",
		},
	)
	if err != nil {
		return err
	}
	blk.CalculateHeadHash()
	blk.CalculateTxsHash()
	blk.CalculateMerkleHash()
	err = bv.BlockChain().Push(blk)
	if err != nil {
		return err
	}
	err = bv.StateDB().Flush(string(blk.HeadHash()))
	if err != nil {
		return err
	}

	return nil
}
