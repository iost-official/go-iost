package verifier

import (
	"encoding/json"
	"io/ioutil"

	"fmt"

	"time"

	"os"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
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
			Number:     0,
			Witness:    "witness",
			Time:       123456,
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

// SetContract without run init
func (s *Simulator) SetContract(c *contract.Contract) {
	s.Visitor.SetContract(c)
}

// DeployContract via iost.system/SetCode
func (s *Simulator) DeployContract(c *contract.Contract, publisher string, kp *account.KeyPair) string {
	trx := tx.NewTx([]*tx.Action{{
		Contract:   "iost.system",
		ActionName: "SetCode",
		Data:       fmt.Sprintf(`["%v"]`, c.B64Encode()),
	}}, nil, int64(100000), int64(1), int64(10000000))

	r, err := s.CallTx(trx, publisher, kp)
	if err != nil {
		panic(err)
	}
	if r.Status.Code != 0 {
		panic(r)
	}
	return "Contract" + common.Base58Encode(trx.Hash())
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
	}}, nil, int64(100000), int64(1), int64(10000000))

	return s.CallTx(trx, publisher, auth)
}

// CallTx with user defiened tx
func (s *Simulator) CallTx(trx *tx.Tx, publisher string, auth *account.KeyPair) (*tx.TxReceipt, error) {
	var sigs = make([]*crypto.Signature, 0)

	//if len(auths) > 1 {
	//	for _, auth := range auths[1:] {
	//		sig, err := tx.SignTxContent(trx, auth)
	//		if err != nil {
	//			return nil, err
	//		}
	//		sigs = append(sigs, sig)
	//	}
	//}

	stx, err := tx.SignTx(trx, publisher, auth, sigs...)
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
		return &tx.TxReceipt{}, err
	}
	r, err := isolator.Run()
	if err != nil {
		return &tx.TxReceipt{}, err
	}
	isolator.Commit()

	return r, nil
}

// Clear mvccdb
func (s *Simulator) Clear() {
	s.mvcc.Close()
	os.RemoveAll("mvcc")
}
