package verifier

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"errors"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm"
	"github.com/iost-official/go-iost/vm/database"
)

// Simulator of txs and contract
type Simulator struct {
	Visitor  *database.Visitor
	Verifier *Verifier
	Head     *block.BlockHead
	Logger   *ilog.Logger
	mvcc     db.MVCCDB
}

// NewSimulator get a simulator with default settings
func NewSimulator() *Simulator {
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		panic(err)
	}
	v := Verifier{}

	return &Simulator{
		Visitor:  database.NewVisitor(0, mvccdb),
		Verifier: &v,
		mvcc:     mvccdb,
		Head: &block.BlockHead{
			ParentHash: []byte("abc"),
			Number:     1,
			Witness:    "witness",
			Time:       int64(1541541540 * 1000 * 1000 * 1000),
		},
		Logger: ilog.DefaultLogger(),
	}
}

// SetBlockHead ...
func (s *Simulator) SetBlockHead(head *block.BlockHead) {
	s.Head = head
}

// SetAccount ...
func (s *Simulator) SetAccount(acc *account.Account) {
	buf, err := json.Marshal(acc)
	if err != nil {
		panic(err)
	}
	s.Visitor.MPut("iost.auth-account", acc.ID, database.MustMarshal(string(buf)))
}

// SetGas to id
func (s *Simulator) SetGas(id string, i int64) {
	s.Visitor.SetGasStock(id, &common.Fixed{
		Value:   i * 10e8,
		Decimal: 8,
	})
	s.Visitor.SetGasLimit(id, &common.Fixed{
		Value:   i * 10e8,
		Decimal: 8,
	})
	s.Visitor.Commit()
}

// SetRAM to id
func (s *Simulator) SetRAM(id string, r int64) {
	s.Visitor.SetTokenBalance("ram", id, r)
	s.Visitor.Commit()
}

// SetContract without run init
func (s *Simulator) SetContract(c *contract.Contract) {
	s.Visitor.SetContract(c)
}

// DeployContract via iost.system/SetCode
func (s *Simulator) DeployContract(c *contract.Contract, publisher string, kp *account.KeyPair) (string, error) {
	trx := tx.NewTx([]*tx.Action{{
		Contract:   "iost.system",
		ActionName: "SetCode",
		Data:       fmt.Sprintf(`["%v"]`, c.B64Encode()),
	}}, nil, 100000, 100, 10000000, 0)

	r, err := s.CallTx(trx, publisher, kp)
	if err != nil {
		return "", err
	}
	if r.Status.Code != 0 {
		return "", errors.New(r.Status.Message)
	}
	return "Contract" + common.Base58Encode(trx.Hash()), nil
}

// Compile files
func (s *Simulator) Compile(id, src, abi string) (*contract.Contract, error) {

	bs, err := ioutil.ReadFile(src + ".js")
	if err != nil {
		return nil, err
	}
	code := string(bs)

	as, err := ioutil.ReadFile(abi + ".abi")
	if err != nil {
		return nil, err
	}

	var info contract.Info
	err = json.Unmarshal(as, &info)
	if err != nil {
		return nil, err
	}
	c := contract.Contract{
		ID:   id,
		Info: &info,
		Code: code,
	}

	return &c, nil
}

// Call abi with basic settings
func (s *Simulator) Call(contractName, abi, args string, publisher string, auth *account.KeyPair) (*tx.TxReceipt, error) {

	trx := tx.NewTx([]*tx.Action{{
		Contract:   contractName,
		ActionName: abi,
		Data:       args,
	}}, nil, 100000, 100, 10000000, 0)

	return s.CallTx(trx, publisher, auth)
}

// CallTx with user defiened tx
func (s *Simulator) CallTx(trx *tx.Tx, publisher string, auth *account.KeyPair) (*tx.TxReceipt, error) {
	//if len(auths) > 1 {
	//	for _, auth := range auths[1:] {
	//		sig, err := tx.SignTxContent(trx, auth)
	//		if err != nil {
	//			return nil, err
	//		}
	//		sigs = append(sigs, sig)
	//	}
	//}

	stx, err := tx.SignTx(trx, publisher, []*account.KeyPair{auth})
	if err != nil {
		return nil, err
	}

	var isolator vm.Isolator

	err = isolator.Prepare(s.Head, s.Visitor, s.Logger)
	if err != nil {
		return &tx.TxReceipt{}, err
	}
	err = isolator.PrepareTx(stx, time.Second)

	if err != nil {
		return &tx.TxReceipt{}, fmt.Errorf("prepare tx error: %v", err)
	}
	_, err = isolator.Run()
	if err != nil {
		return &tx.TxReceipt{}, err
	}

	r, err := isolator.PayCost()
	if err != nil {
		return nil, err
	}
	isolator.Commit()

	return r, nil
}

// Clear mvccdb
func (s *Simulator) Clear() {
	s.mvcc.Close()
	os.RemoveAll("mvcc")
}
