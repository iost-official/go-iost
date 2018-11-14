package genesis

import (
	"encoding/json"
	"fmt"
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

func compile(id string, path string, name string) (*contract.Contract, error) {
	if id == "" || path == "" || name == "" {
		return nil, fmt.Errorf("arguments is error, id:%v, path:%v, name:%v", id, path, name)
	}
	cFilePath := filepath.Join(path, name)
	cAbiPath := filepath.Join(path, name+".abi")
	return contract.Compile(id, cFilePath, cAbiPath)
}

func genGenesisTx(gConf *common.GenesisConfig) (*tx.Tx, *account.KeyPair, error) {
	witnessInfo := gConf.WitnessInfo
	// new account
	keyPair, err := account.NewKeyPair(common.Base58Decode("2vj2Ab8Taz1TT2MSQHxmSffGnvsc9EVrmjx1W7SBQthCpuykhbRn2it8DgNkcm4T9tdBgsue3uBiAzxLpLJoDUbc"), crypto.Ed25519)
	if err != nil {
		return nil, nil, err
	}

	// prepare actions
	var acts []*tx.Action

	// deploy iost.account
	code, err := compile("iost.auth", gConf.ContractPath, "account.js")
	if err != nil {
		return nil, nil, err
	}
	acts = append(acts, tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.auth", code.B64Encode())))

	// new account
	adminInfo := gConf.AdminInfo
	acts = append(acts, tx.NewAction("iost.auth", "SignUp", fmt.Sprintf(`["%v", "%v", "%v"]`, adminInfo.ID, adminInfo.Owner, adminInfo.Active)))
	// new account
	foundationInfo := gConf.FoundationInfo
	acts = append(acts, tx.NewAction("iost.auth", "SignUp", fmt.Sprintf(`["%v", "%v", "%v"]`, foundationInfo.ID, foundationInfo.Owner, foundationInfo.Active)))
	// init account
	acts = append(acts, tx.NewAction("iost.auth", "SignUp", fmt.Sprintf(`["%v", "%v", "%v"]`, "inituser", keyPair.ID, keyPair.ID)))

	for _, v := range witnessInfo {
		acts = append(acts, tx.NewAction("iost.auth", "SignUp", fmt.Sprintf(`["%v", "%v", "%v"]`, v.ID, v.Owner, v.Active)))
	}

	// deploy iost.token
	acts = append(acts, tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.token", native.TokenABI().B64Encode())))

	// deploy iost.bonus
	code, err = compile("iost.bonus", gConf.ContractPath, "bonus.js")
	if err != nil {
		return nil, nil, err
	}
	acts = append(acts, tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.bonus", code.B64Encode())))
	acts = append(acts, tx.NewAction("iost.bonus", "InitAdmin", fmt.Sprintf(`["%v"]`, adminInfo.ID)))

	// deploy iost.issue and create iost
	code, err = compile("iost.issue", gConf.ContractPath, "issue.js")
	if err != nil {
		return nil, nil, err
	}
	acts = append(acts, tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.issue", code.B64Encode())))
	genesisConfig := gConf.TokenInfo
	tokenHolder := append(witnessInfo, adminInfo)
	params := []interface{}{
		adminInfo.ID,
		genesisConfig,
		tokenHolder,
	}
	b, _ := json.Marshal(params)
	acts = append(acts, tx.NewAction("iost.issue", "InitGenesis", string(b)))

	// deploy iost.vote
	code, err = compile("iost.vote", gConf.ContractPath, "vote_common.js")
	if err != nil {
		return nil, nil, err
	}
	acts = append(acts, tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.vote", code.B64Encode())))
	acts = append(acts, tx.NewAction("iost.vote", "InitAdmin", fmt.Sprintf(`["%v"]`, adminInfo.ID)))

	// deploy iost.vote_producer
	code, err = compile("iost.vote_producer", gConf.ContractPath, "vote.js")
	if err != nil {
		return nil, nil, err
	}
	acts = append(acts, tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.vote_producer", code.B64Encode())))

	// deploy iost.base
	code, err = compile("iost.base", gConf.ContractPath, "base.js")
	if err != nil {
		return nil, nil, err
	}
	acts = append(acts, tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.base", code.B64Encode())))

	for _, v := range witnessInfo {
		acts = append(acts, tx.NewAction("iost.vote_producer", "InitProducer", fmt.Sprintf(`["%v", "%v"]`, v.ID, v.Active)))
	}
	acts = append(acts, tx.NewAction("iost.vote_producer", "InitAdmin", fmt.Sprintf(`["%v"]`, adminInfo.ID)))

	// deploy iost.gas
	acts = append(acts, tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.gas", native.GasABI().B64Encode())))

	// pledge gas for admin
	gasPledgeAmount := 100000000
	acts = append(acts, tx.NewAction("iost.gas", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, adminInfo.ID, adminInfo.ID, gasPledgeAmount)))

	// deploy iost.ram
	code, err = compile("iost.ram", gConf.ContractPath, "ram.js")
	if err != nil {
		return nil, nil, err
	}
	acts = append(acts, tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.ram", code.B64Encode())))
	acts = append(acts, tx.NewAction("iost.ram", "initAdmin", fmt.Sprintf(`["%v"]`, adminInfo.ID)))
	acts = append(acts, tx.NewAction("iost.ram", "initContractName", fmt.Sprintf(`["%v"]`, "iost.ram")))
	var initialTotal int64 = 128 * 1024 * 1024 * 1024        // 128GB at first
	var increaseInterval int64 = 24 * 3600 / 3               // increase every day
	var increaseAmount int64 = 64 * 1024 * 1024 * 1024 / 365 // 64GB per year
	acts = append(acts, tx.NewAction("iost.ram", "issue", fmt.Sprintf(`[%v, %v, %v]`, initialTotal, increaseInterval, increaseAmount)))
	adminInitialRAM := 1024
	acts = append(acts, tx.NewAction("iost.ram", "buy", fmt.Sprintf(`["%v", "%v", %v]`, adminInfo.ID, adminInfo.ID, adminInitialRAM)))

	trx := tx.NewTx(acts, nil, 100000000, 0, 0, 0)
	trx.Time = 0
	trx, err = tx.SignTx(trx, "inituser@active", []*account.KeyPair{keyPair})
	if err != nil {
		return nil, nil, err
	}
	return trx, keyPair, nil
}

// GenGenesis is create a genesis block
func GenGenesis(db db.MVCCDB, gConf *common.GenesisConfig) (*block.Block, error) {
	t, err := time.Parse(time.RFC3339, gConf.InitialTimestamp)
	if err != nil {
		ilog.Fatalf("invalid genesis initial time string %v (%v).", gConf.InitialTimestamp, err)
		return nil, err
	}
	trx, acc, err := genGenesisTx(gConf)
	if err != nil {
		return nil, err
	}

	blockHead := block.BlockHead{
		Version:    0,
		ParentHash: nil,
		Number:     0,
		Witness:    acc.ID,
		Time:       t.UnixNano(),
		GasUsage:   0,
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
	blk.Head.GasUsage = txr.GasUsage
	blk.Head.TxsHash = blk.CalculateTxsHash()
	blk.Head.MerkleHash = blk.CalculateMerkleHash()
	if err != nil {
		return nil, err
	}
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
			ContractPath:     os.Getenv("GOPATH") + "/src/github.com/iost-official/go-iost/config/",
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
