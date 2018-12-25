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
	"github.com/iost-official/go-iost/vm/native"
)

// GenesisTxExecTime is the maximum execution time of a transaction in genesis block
var GenesisTxExecTime = 3 * time.Second

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

	// deploy account.iost
	code, err := compile("auth.iost", gConf.ContractPath, "account.js")
	if err != nil {
		return nil, nil, err
	}
	acts = append(acts, tx.NewAction("system.iost", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "auth.iost", code.B64Encode())))

	// deploy iost.gas
	acts = append(acts, tx.NewAction("system.iost", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "gas.iost", native.GasABI().B64Encode())))

	// new account
	adminInfo := gConf.AdminInfo
	acts = append(acts, tx.NewAction("auth.iost", "SignUp", fmt.Sprintf(`["%v", "%v", "%v"]`, adminInfo.ID, adminInfo.Owner, adminInfo.Active)))
	// new account
	foundationInfo := gConf.FoundationInfo
	acts = append(acts, tx.NewAction("auth.iost", "SignUp", fmt.Sprintf(`["%v", "%v", "%v"]`, foundationInfo.ID, foundationInfo.Owner, foundationInfo.Active)))
	// init account
	acts = append(acts, tx.NewAction("auth.iost", "SignUp", fmt.Sprintf(`["%v", "%v", "%v"]`, "inituser", keyPair.ID, keyPair.ID)))

	for _, v := range witnessInfo {
		acts = append(acts, tx.NewAction("auth.iost", "SignUp", fmt.Sprintf(`["%v", "%v", "%v"]`, v.ID, v.Owner, v.Active)))
	}
	invalidPubKey := "11111111111111111111111111111111"
	acts = append(acts, tx.NewAction("auth.iost", "SignUp", fmt.Sprintf(`["%v", "%v", "%v"]`, "deadaddr", invalidPubKey, invalidPubKey)))

	// deploy token.iost
	acts = append(acts, tx.NewAction("system.iost", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "token.iost", native.TokenABI().B64Encode())))
	acts = append(acts, tx.NewAction("system.iost", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "token721.iost", native.Token721ABI().B64Encode())))

	// deploy bonus.iost
	code, err = compile("bonus.iost", gConf.ContractPath, "bonus.js")
	if err != nil {
		return nil, nil, err
	}
	acts = append(acts, tx.NewAction("system.iost", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "bonus.iost", code.B64Encode())))
	acts = append(acts, tx.NewAction("bonus.iost", "InitAdmin", fmt.Sprintf(`["%v"]`, adminInfo.ID)))

	// deploy issue.iost and create iost
	code, err = compile("issue.iost", gConf.ContractPath, "issue.js")
	if err != nil {
		return nil, nil, err
	}
	acts = append(acts, tx.NewAction("system.iost", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "issue.iost", code.B64Encode())))
	genesisConfig := gConf.TokenInfo
	tokenHolder := append(witnessInfo, adminInfo)
	params := []interface{}{
		adminInfo.ID,
		genesisConfig,
		tokenHolder,
	}
	b, _ := json.Marshal(params)
	acts = append(acts, tx.NewAction("issue.iost", "InitGenesis", string(b)))

	// deploy vote.iost
	code, err = compile("vote.iost", gConf.ContractPath, "vote_common.js")
	if err != nil {
		return nil, nil, err
	}
	acts = append(acts, tx.NewAction("system.iost", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "vote.iost", code.B64Encode())))
	acts = append(acts, tx.NewAction("vote.iost", "InitAdmin", fmt.Sprintf(`["%v"]`, adminInfo.ID)))

	// deploy vote_producer.iost
	code, err = compile("vote_producer.iost", gConf.ContractPath, "vote.js")
	if err != nil {
		return nil, nil, err
	}
	acts = append(acts, tx.NewAction("system.iost", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "vote_producer.iost", code.B64Encode())))

	// deploy base.iost
	code, err = compile("base.iost", gConf.ContractPath, "base.js")
	if err != nil {
		return nil, nil, err
	}
	acts = append(acts, tx.NewAction("system.iost", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "base.iost", code.B64Encode())))
	acts = append(acts, tx.NewAction("base.iost", "InitAdmin", fmt.Sprintf(`["%v"]`, adminInfo.ID)))

	for _, v := range witnessInfo {
		acts = append(acts, tx.NewAction("vote_producer.iost", "InitProducer", fmt.Sprintf(`["%v", "%v"]`, v.ID, v.Active)))
	}
	acts = append(acts, tx.NewAction("vote_producer.iost", "InitAdmin", fmt.Sprintf(`["%v"]`, adminInfo.ID)))

	// pledge gas for admin
	gasPledgeAmount := 100
	acts = append(acts, tx.NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, adminInfo.ID, adminInfo.ID, gasPledgeAmount)))

	// deploy ram.iost
	code, err = compile("ram.iost", gConf.ContractPath, "ram.js")
	if err != nil {
		return nil, nil, err
	}
	acts = append(acts, tx.NewAction("system.iost", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "ram.iost", code.B64Encode())))
	acts = append(acts, tx.NewAction("ram.iost", "initAdmin", fmt.Sprintf(`["%v"]`, adminInfo.ID)))
	var initialTotal int64 = 128 * 1024 * 1024 * 1024                           // 128GB at first
	var increaseInterval int64 = 10 * 60                                        // increase every 10 mins
	var increaseAmount int64 = 10 * (64 * 1024 * 1024 * 1024) / (365 * 24 * 60) // 64GB per year
	var reserveRAM = initialTotal * 3 / 10                                      // reserve for foundation
	acts = append(acts, tx.NewAction("ram.iost", "issue", fmt.Sprintf(`[%v, %v, %v, %v]`, initialTotal, increaseInterval, increaseAmount, reserveRAM)))

	adminInitialRAM := 100000
	acts = append(acts, tx.NewAction("ram.iost", "buy", fmt.Sprintf(`["%v", "%v", %v]`, adminInfo.ID, adminInfo.ID, adminInitialRAM)))
	acts = append(acts, tx.NewAction("token.iost", "transfer", fmt.Sprintf(`["ram","ram.iost", "%v", "%v", ""]`, foundationInfo.ID, reserveRAM)))

	for _, v := range witnessInfo {
		acts = append(acts, tx.NewAction("ram.iost", "buy", fmt.Sprintf(`["%v", "%v", %v]`, adminInfo.ID, v.ID, adminInitialRAM)))
	}

	acts = append(acts, tx.NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, adminInfo.ID, foundationInfo.ID, gasPledgeAmount)))
	for _, v := range witnessInfo {
		acts = append(acts, tx.NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, adminInfo.ID, v.ID, gasPledgeAmount)))
	}

	trx := tx.NewTx(acts, nil, 1000000000, 100, 0, 0)
	trx.Time = 0
	trx, err = tx.SignTx(trx, "inituser", []*account.KeyPair{keyPair})
	if err != nil {
		return nil, nil, err
	}
	trx.AmountLimit = append(trx.AmountLimit, &contract.Amount{Token: "*", Val: "unlimited"})
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
	blk.Head.TxMerkleHash = blk.CalculateTxMerkleHash()
	blk.Head.TxReceiptMerkleHash = blk.CalculateTxReceiptMerkleHash()
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
	blk.CalculateTxMerkleHash()
	blk.CalculateTxReceiptMerkleHash()
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
