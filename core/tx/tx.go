package tx

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx/pb"
	"github.com/iost-official/go-iost/crypto"
)

//go:generate protoc  --go_out=plugins=grpc:. ./core/tx/tx.proto

// Tx Transaction structure
type Tx struct {
	hash         []byte
	Time         int64               `json:"time"`
	Expiration   int64               `json:"expiration"`
	GasPrice     int64               `json:"gas_price"`
	GasLimit     int64               `json:"gas_limit"`
	Delay        int64               `json:"delay"`
	Actions      []*Action           `json:"-"`
	Signers      []string            `json:"-"`
	Signs        []*crypto.Signature `json:"-"`
	Publisher    string              `json:"-"`
	PublishSigns []*crypto.Signature `json:"-"`
	ReferredTx   []byte              `json:"referred_tx"`
}

// NewTx return a new Tx
func NewTx(actions []*Action, signers []string, gasLimit, gasPrice, expiration, delay int64) *Tx {
	return &Tx{
		Time:         time.Now().UnixNano(),
		Actions:      actions,
		Signers:      signers,
		GasLimit:     gasLimit,
		GasPrice:     gasPrice,
		Expiration:   expiration,
		hash:         nil,
		PublishSigns: []*crypto.Signature{},
		Delay:        delay,
	}
}

// SignTxContent sign tx content, only signers should do this
func SignTxContent(tx *Tx, id string, account *account.KeyPair) (*crypto.Signature, error) {
	if !tx.containSigner(id) {
		return nil, errors.New("account not included in signer list of this transaction")
	}
	return account.Sign(tx.baseHash()), nil
}

func (t *Tx) containSigner(id string) bool {
	for _, signer := range t.Signers {
		if strings.HasPrefix(signer, id) {
			return true
		}
	}
	return false
}

func (t *Tx) baseHash() []byte {
	tr := &txpb.Tx{
		Time:       t.Time,
		Expiration: t.Expiration,
		GasLimit:   t.GasLimit,
		GasPrice:   t.GasPrice,
		Signers:    t.Signers,
		Delay:      t.Delay,
	}
	for _, a := range t.Actions {
		tr.Actions = append(tr.Actions, a.ToPb())
	}

	b, err := tr.Marshal()
	if err != nil {
		panic(err)
	}
	return common.Sha3(b)
}

// SignTx sign the whole tx, including signers' signature, only publisher should do this
func SignTx(tx *Tx, id string, kps []*account.KeyPair, signs ...*crypto.Signature) (*Tx, error) {
	tx.Signs = append(tx.Signs, signs...)

	tx.PublishSigns = []*crypto.Signature{}
	for _, kp := range kps {
		sig := kp.Sign(tx.publishHash())
		tx.PublishSigns = append(tx.PublishSigns, sig)
	}
	tx.Publisher = id
	tx.hash = nil
	return tx, nil
}

// publishHash
func (t *Tx) publishHash() []byte {
	tr := &txpb.Tx{
		Time:       t.Time,
		Expiration: t.Expiration,
		GasLimit:   t.GasLimit,
		GasPrice:   t.GasPrice,
		Signers:    t.Signers,
		Delay:      t.Delay,
	}
	for _, a := range t.Actions {
		tr.Actions = append(tr.Actions, a.ToPb())
	}

	for _, s := range t.Signs {
		tr.Signs = append(tr.Signs, s.ToPb())
	}

	b, err := tr.Marshal()
	if err != nil {
		panic(err)
	}
	return common.Sha3(b)
}

// ToPb convert tx to txpb.Tx for transmission.
func (t *Tx) ToPb() *txpb.Tx {
	tr := &txpb.Tx{
		Time:       t.Time,
		Expiration: t.Expiration,
		GasLimit:   t.GasLimit,
		GasPrice:   t.GasPrice,
		Signers:    t.Signers,
		Delay:      t.Delay,
		ReferredTx: t.ReferredTx,
	}
	for _, a := range t.Actions {
		tr.Actions = append(tr.Actions, a.ToPb())
	}

	for _, s := range t.Signs {
		tr.Signs = append(tr.Signs, s.ToPb())
	}
	tr.Publisher = t.Publisher
	for _, sig := range t.PublishSigns {
		tr.PublishSigns = append(tr.PublishSigns, sig.ToPb())
	}
	return tr
}

// Encode tx to byte array
func (t *Tx) Encode() []byte {
	tr := t.ToPb()
	b, err := tr.Marshal()
	if err != nil {
		panic(err)
	}
	return b
}

// FromPb convert tx from txpb.Tx.
func (t *Tx) FromPb(tr *txpb.Tx) *Tx {
	t.Time = tr.Time
	t.Expiration = tr.Expiration
	t.GasLimit = tr.GasLimit
	t.GasPrice = tr.GasPrice
	t.Actions = []*Action{}
	t.Delay = tr.Delay
	t.ReferredTx = tr.ReferredTx
	for _, a := range tr.Actions {
		ac := &Action{}
		t.Actions = append(t.Actions, ac.FromPb(a))
	}
	t.Signers = tr.Signers
	t.Signs = []*crypto.Signature{}
	for _, sr := range tr.Signs {
		sig := &crypto.Signature{}
		t.Signs = append(t.Signs, sig.FromPb(sr))
	}
	t.Publisher = tr.Publisher
	t.PublishSigns = []*crypto.Signature{}
	for _, sr := range tr.PublishSigns {
		sig := &crypto.Signature{}
		t.PublishSigns = append(t.PublishSigns, sig.FromPb(sr))
	}
	t.hash = nil
	return t
}

// Decode tx from byte array
func (t *Tx) Decode(b []byte) error {
	tr := &txpb.Tx{}
	err := tr.Unmarshal(b)
	if err != nil {
		return err
	}
	t.FromPb(tr)
	return nil
}

// String return human-readable tx
func (t *Tx) String() string {
	str := "Tx{\n"
	str += "	Time: " + strconv.FormatInt(t.Time, 10) + ",\n"
	str += "	Publisher: " + t.Publisher + ",\n"
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

// IsDefer returns whether the transaction is a defer tx.
//
// Defer transaction is the transaction that's generated by a delay tx.
func (t *Tx) IsDefer() bool {
	return len(t.ReferredTx) > 0
}

// VerifySelf verify tx's signature
func (t *Tx) VerifySelf() error { // only check whether sigs are legal
	if t.Delay > 0 && t.IsDefer() {
		return errors.New("invalid tx. including both delaysecond and referredtx")
	}
	if t.IsDefer() {
		return nil
	}
	baseHash := t.baseHash()
	//signerSet := make(map[string]bool)
	for _, sign := range t.Signs {
		ok := sign.Verify(baseHash)
		if !ok {
			return fmt.Errorf("signer error")
		}
		//signerSet[account.GetIDByPubkey(sign.Pubkey)] = true
	}
	//for _, signer := range t.Signers {
	//	if _, ok := signerSet[signer]; !ok {
	//		return fmt.Errorf("signer not enough")
	//	}
	//}
	if len(t.PublishSigns) == 0 {
		return fmt.Errorf("publisher empty error")
	}
	for _, sign := range t.PublishSigns {
		ok := sign != nil && sign.Verify(t.publishHash())
		if !ok {
			return fmt.Errorf("publisher error")
		}
	}
	return nil
}

// VerifySigner verify signer's signature
func (t *Tx) VerifySigner(sig *crypto.Signature) bool {
	return sig.Verify(t.baseHash())
}
