package genesis

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"encoding/json"
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
	"io/ioutil"
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

func Compile(id, src, abi string) (*contract.Contract, error) {
	println(id, src, abi)

	bs, err := ioutil.ReadFile(src)
	if err != nil {
		return nil, err
	}
	code := string(bs)

	as, err := ioutil.ReadFile(abi)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(as))
	var info contract.Info
	err = json.Unmarshal(as, &info)
	if err != nil {
		return nil, err
	}
	fmt.Println(info)
	c := contract.Contract{
		ID:   id,
		Info: &info,
		Code: code,
	}

	return &c, nil
}

func genGenesisTx(gConf *common.GenesisConfig) (*tx.Tx, *account.KeyPair, error) {
	witnessInfo := gConf.WitnessInfo
	// new account
	acc, err := account.NewKeyPair(common.Base58Decode("2vj2Ab8Taz1TT2MSQHxmSffGnvsc9EVrmjx1W7SBQthCpuykhbRn2it8DgNkcm4T9tdBgsue3uBiAzxLpLJoDUbc"), crypto.Ed25519)
	if err != nil {
		return nil, nil, err
	}

	// prepare actions
	var acts []*tx.Action

	// deploy iost.account
	accountFilePath := filepath.Join(gConf.ContractPath, "account.js")
	accountAbiPath := filepath.Join(gConf.ContractPath, "account.js.abi")
	code, err := Compile("iost.auth", accountFilePath, accountAbiPath)

	if err != nil {
		return nil, nil, err
	}
	acts = append(acts, tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.auth", code.B64Encode())))

	// new account
	adminInfo := gConf.AdminInfo
	acts = append(acts, tx.NewAction("iost.auth", "SignUp", fmt.Sprintf(`["%v", "%v", "%v"]`, adminInfo.ID, adminInfo.Owner, adminInfo.Active)))
	// init account
	acts = append(acts, tx.NewAction("iost.auth", "SignUp", fmt.Sprintf(`["%v", "%v", "%v"]`, acc.ID, acc.ID, acc.ID)))

	for _, v := range witnessInfo {
		acts = append(acts, tx.NewAction("iost.auth", "SignUp", fmt.Sprintf(`["%v", "%v", "%v"]`, v.ID, v.Owner, v.Active)))
	}

	// deploy iost.token and create iost
	acts = append(acts, tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.token", native.TokenABI().B64Encode())))
	acts = append(acts, tx.NewAction("iost.token", "create", fmt.Sprintf(`["iost", "%v", 21000000000, {}]`, acc.ID)))

	// issue token
	for _, v := range witnessInfo {
		acts = append(acts, tx.NewAction("iost.token", "issue", fmt.Sprintf(`["iost", "%v", "%v"]`, v.ID, v.Balance)))
	}
	acts = append(acts, tx.NewAction("iost.token", "issue", fmt.Sprintf(`["iost", "%v", "%v"]`, acc.ID, adminInfo.Balance)))

	// deploy iost.vote
	voteFilePath := filepath.Join(gConf.ContractPath, "vote.js")
	voteAbiPath := filepath.Join(gConf.ContractPath, "vote.js.abi")
	code, err = Compile("iost.vote", voteFilePath, voteAbiPath)
	if err != nil {
		return nil, nil, err
	}
	acts = append(acts, tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.vote", code.B64Encode())))

	for _, v := range witnessInfo {
		acts = append(acts, tx.NewAction("iost.vote", "InitProducer", fmt.Sprintf(`["%v"]`, v.Owner)))
	}
	acts = append(acts, tx.NewAction("iost.vote", "InitAdmin", fmt.Sprintf(`["%v"]`, adminInfo.ID)))

	// deploy iost.bonus
	acts = append(acts, tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.bonus", native.BonusABI().B64Encode())))

	// deploy iost.gas
	acts = append(acts, tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.gas", native.GasABI().B64Encode())))

	// todo pledge gas for admin
	// acts = append(acts, tx.NewAction("iost.gas", "PledgeGas", fmt.Sprintf(`["%v", "%v"]`, adminInfo.ID, adminInfo.Balance)))

	trx := tx.NewTx(acts, nil, 100000000, 0, 0, 0)
	trx.Time = 0
	trx, err = tx.SignTx(trx, acc.ID, []*account.KeyPair{acc})
	if err != nil {
		return nil, nil, err
	}
	return trx, acc, nil
}

// GenGenesis is create a genesis block
func GenGenesis(db db.MVCCDB, gConf *common.GenesisConfig) (*block.Block, error) {
	t, err := common.ParseStringToTimestamp(gConf.InitialTimestamp)
	if err != nil {
		ilog.Fatalf("invalid genesis initial time string %v (%v).", gConf.InitialTimestamp, err)
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
