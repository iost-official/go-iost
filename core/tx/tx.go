package tx

import (
	"fmt"
	"time"

	"bytes"
	"errors"
	"strconv"

	"github.com/gogo/protobuf/proto"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto"
)

//go:generate protoc  --go_out=plugins=grpc:. ./core/tx/tx.proto

// Tx Transaction structure
type Tx struct {
	hash        []byte
	Time        int64               `json:"time,string"`
	Expiration  int64               `json:"expiration,string"`
	GasLimit    int64               `json:"gas_limit,string"`
	Actions     []*Action           `json:"-"`
	Signers     [][]byte            `json:"-"`
	Signs       []*crypto.Signature `json:"-"`
	Publisher   string              `json:"-"`
	PublishSign *crypto.Signature   `json:"-"`
	GasPrice    int64               `json:"gas_price,string"`
}

// NewTx return a new Tx
func NewTx(actions []*Action, signers [][]byte, gasLimit int64, gasPrice int64, expiration int64) *Tx {
	now := time.Now().UnixNano()
	return &Tx{
		Time:        now,
		Actions:     actions,
		Signers:     signers,
		GasLimit:    gasLimit,
		GasPrice:    gasPrice,
		Expiration:  expiration,
		hash:        nil,
		PublishSign: &crypto.Signature{},
	}
}

// SignTxContent sign tx content, only signers should do this
func SignTxContent(tx *Tx, account *account.KeyPair) (*crypto.Signature, error) {
	if !tx.containSigner(account.Pubkey) {
		return nil, errors.New("account not included in signer list of this transaction")
	}
	return account.Sign(tx.baseHash()), nil
}
func (t *Tx) containSigner(pubkey []byte) bool {
	found := false
	for _, signer := range t.Signers {
		if bytes.Equal(signer, pubkey) {
			found = true
		}
	}
	return found
}

func (t *Tx) baseHash() []byte {
	tr := &TxRaw{
		Time:       t.Time,
		Expiration: t.Expiration,
		GasLimit:   t.GasLimit,
		GasPrice:   t.GasPrice,
	}
	for _, a := range t.Actions {
		tr.Actions = append(tr.Actions, &ActionRaw{
			Contract:   a.Contract,
			ActionName: a.ActionName,
			Data:       a.Data,
		})
	}
	tr.Signers = t.Signers

	b, err := proto.Marshal(tr)
	if err != nil {
		panic(err)
	}
	return common.Sha3(b)
}

// SignTx sign the whole tx, including signers' signature, only publisher should do this
func SignTx(tx *Tx, id string, kp *account.KeyPair, signs ...*crypto.Signature) (*Tx, error) {
	tx.Signs = append(tx.Signs, signs...)

	sig := kp.Sign(tx.publishHash())
	tx.PublishSign = sig
	tx.Publisher = id
	tx.hash = nil
	return tx, nil
}

// publishHash
func (t *Tx) publishHash() []byte {
	tr := &TxRaw{
		Time:       t.Time,
		Expiration: t.Expiration,
		GasLimit:   t.GasLimit,
		GasPrice:   t.GasPrice,
	}
	for _, a := range t.Actions {
		tr.Actions = append(tr.Actions, &ActionRaw{
			Contract:   a.Contract,
			ActionName: a.ActionName,
			Data:       a.Data,
		})
	}
	tr.Signers = t.Signers
	for _, s := range t.Signs {
		tr.Signs = append(tr.Signs, &crypto.SignatureRaw{
			Algorithm: int32(s.Algorithm),
			Sig:       s.Sig,
			PubKey:    s.Pubkey,
		})
	}

	b, err := proto.Marshal(tr)
	if err != nil {
		panic(err)
	}
	return common.Sha3(b)
}

// ToTxRaw convert tx to TxRaw for transmission
func (t *Tx) ToTxRaw() *TxRaw {
	tr := &TxRaw{
		Time:       t.Time,
		Expiration: t.Expiration,
		GasLimit:   t.GasLimit,
		GasPrice:   t.GasPrice,
	}
	for _, a := range t.Actions {
		tr.Actions = append(tr.Actions, &ActionRaw{
			Contract:   a.Contract,
			ActionName: a.ActionName,
			Data:       a.Data,
		})
	}
	tr.Signers = t.Signers
	for _, s := range t.Signs {
		tr.Signs = append(tr.Signs, &crypto.SignatureRaw{
			Algorithm: int32(s.Algorithm),
			Sig:       s.Sig,
			PubKey:    s.Pubkey,
		})
	}
	tr.Publisher = t.Publisher
	if t.PublishSign != nil {
		tr.PublishSign = &crypto.SignatureRaw{
			Algorithm: int32(t.PublishSign.Algorithm),
			Sig:       t.PublishSign.Sig,
			PubKey:    t.PublishSign.Pubkey,
		}
	}
	return tr
}

// Encode tx to byte array
func (t *Tx) Encode() []byte {
	tr := t.ToTxRaw()
	b, err := proto.Marshal(tr)
	if err != nil {
		panic(err)
	}
	return b
}

// FromTxRaw convert tx from TxRaw
func (t *Tx) FromTxRaw(tr *TxRaw) {
	t.Time = tr.Time
	t.Expiration = tr.Expiration
	t.GasLimit = tr.GasLimit
	t.GasPrice = tr.GasPrice
	t.Actions = []*Action{}
	for _, a := range tr.Actions {
		t.Actions = append(t.Actions, &Action{
			Contract:   a.Contract,
			ActionName: a.ActionName,
			Data:       a.Data,
		})
	}
	t.Signers = tr.Signers
	t.Signs = []*crypto.Signature{}
	for _, sr := range tr.Signs {
		t.Signs = append(t.Signs, &crypto.Signature{
			Algorithm: crypto.Algorithm(sr.Algorithm),
			Sig:       sr.Sig,
			Pubkey:    sr.PubKey,
		})
	}
	t.Publisher = tr.Publisher
	if tr.PublishSign != nil {
		t.PublishSign = &crypto.Signature{
			Algorithm: crypto.Algorithm(tr.PublishSign.Algorithm),
			Sig:       tr.PublishSign.Sig,
			Pubkey:    tr.PublishSign.PubKey,
		}
	}
	t.hash = nil
}

// Decode tx from byte array
func (t *Tx) Decode(b []byte) error {
	tr := &TxRaw{}
	err := proto.Unmarshal(b, tr)
	if err != nil {
		return err
	}
	t.FromTxRaw(tr)
	return nil
}

// String return human-readable tx
func (t *Tx) String() string {
	str := "Tx{\n"
	str += "	Time: " + strconv.FormatInt(t.Time, 10) + ",\n"
	str += "	Pubkey: " + string(t.PublishSign.Pubkey) + ",\n"
	str += "	Action:\n"
	for _, a := range t.Actions {
		str += "		" + a.String()
	}
	str += "}\n"
	return str
}

// Hash return cached hash if exists, or calculate with Sha3
func (t *Tx) Hash() []byte {
	if t.hash == nil {
		t.hash = common.Sha3(t.Encode())
	}
	return t.hash
}

// VerifySelf verify tx's signature
func (t *Tx) VerifySelf() error {
	baseHash := t.baseHash()
	signerSet := make(map[string]bool)
	for _, sign := range t.Signs {
		ok := sign.Verify(baseHash)
		if !ok {
			return fmt.Errorf("signer error")
		}
		signerSet[string(sign.Pubkey)] = true
	}
	for _, signer := range t.Signers {
		if _, ok := signerSet[string(signer)]; !ok {
			return fmt.Errorf("signer not enough")
		}
	}
	ok := t.PublishSign != nil && t.PublishSign.Verify(t.publishHash())
	if !ok {
		return fmt.Errorf("publisher error")
	}
	return nil
}

// VerifySigner verify signer's signature
func (t *Tx) VerifySigner(sig *crypto.Signature) bool {
	return sig.Verify(t.baseHash())
}
