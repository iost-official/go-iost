package verifier

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

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
	Mvcc     db.MVCCDB
}

// NewSimulator get a simulator with default settings
func NewSimulator() *Simulator {
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		panic(err)
	}
	v := Verifier{}

	s := &Simulator{
		Visitor:  database.NewVisitor(0, mvccdb),
		Verifier: &v,
		Mvcc:     mvccdb,
		Head: &block.BlockHead{
			ParentHash: []byte("abc"),
			Number:     1,
			Witness:    "witness",
			Time:       int64(1541541540 * 1000 * 1000 * 1000),
		},
		Logger: ilog.DefaultLogger(),
	}
	return s
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
	s.Visitor.MPut("auth.iost"+"-auth", acc.ID, database.MustMarshal(string(buf)))
}

// SetGas to id
func (s *Simulator) SetGas(id string, i int64) {
	prefix := database.GasContractName + database.Separator
	value := &common.Fixed{
		Value:   i * 10e2,
		Decimal: 2,
	}
	valueStr := database.MustMarshal(value)
	s.Visitor.Put(prefix+database.GasStockKey+id, valueStr)
	s.Visitor.Put(prefix+database.GasLimitKey+id, valueStr)
	s.Visitor.Commit()
}

// SetRAM to id
func (s *Simulator) SetRAM(id string, r int64) {
	s.Visitor.SetTokenBalance("ram", id, r)
	s.Visitor.Commit()
}

// GetRAM of id
func (s *Simulator) GetRAM(id string) int64 {
	return s.Visitor.TokenBalance("ram", id)
}

// SetContract without run init
func (s *Simulator) SetContract(c *contract.Contract) {
	s.Visitor.SetContract(c)
}

// DeployContract via system.iost/SetCode
func (s *Simulator) DeployContract(c *contract.Contract, publisher string, kp *account.KeyPair) (string, *tx.TxReceipt, error) {
	sc, err := json.Marshal(c)
	if err != nil {
		return "", nil, nil
	}

	jargs, err := json.Marshal([]string{string(sc)})
	if err != nil {
		panic(err)
	}

	trx := tx.NewTx([]*tx.Action{{
		Contract:   "system.iost",
		ActionName: "SetCode",
		Data:       string(jargs),
	}}, nil, 10000000, 100, s.Head.Time+10000000, 0)

	trx.Time = s.Head.Time

	r, err := s.CallTx(trx, publisher, kp)
	if err != nil {
		return "", r, err
	}
	if r.Status.Code != 0 {
		return "", r, errors.New(r.Status.Message)
	}
	return "Contract" + common.Base58Encode(trx.Hash()), r, nil
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
	}}, nil, 100000000, 100, s.Head.Time+10000000, 0)

	trx.Time = s.Head.Time

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
	s.Mvcc.Close()
	os.RemoveAll("mvcc")
}
