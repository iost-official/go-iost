package tx

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx/pb"
	"github.com/iost-official/go-iost/crypto"
)

const (
	minGasRatio = 100
	maxGasRatio = 10000
	minGasLimit = 500000
	maxGasLimit = 200000000
	txSizeLimit = 65536
)

// values
var (
	MaxExpiration = int64(90 * time.Second)
)

//go:generate protoc  --go_out=plugins=grpc:. ./core/tx/tx.proto

// ToBytesLevel judges which fields of tx should be written to bytes.
type ToBytesLevel int

// consts
const (
	Base ToBytesLevel = iota
	Publish
	Full
)

// Tx Transaction structure
type Tx struct {
	hash         []byte
	Time         int64               `json:"time"`
	Expiration   int64               `json:"expiration"`
	GasRatio     int64               `json:"gas_ratio"`
	GasLimit     int64               `json:"gas_limit"`
	Delay        int64               `json:"delay"`
	Actions      []*Action           `json:"-"`
	Signers      []string            `json:"-"`
	Signs        []*crypto.Signature `json:"-"`
	Publisher    string              `json:"-"`
	PublishSigns []*crypto.Signature `json:"-"`
	ReferredTx   []byte              `json:"referred_tx"`
	AmountLimit  []*contract.Amount  `json:"amountLimit"`
}

// NewTx return a new Tx
func NewTx(actions []*Action, signers []string, gasLimit, gasRatio, expiration, delay int64) *Tx {
	return &Tx{
		Time:         time.Now().UnixNano(),
		Actions:      actions,
		Signers:      signers,
		GasLimit:     gasLimit,
		GasRatio:     gasRatio,
		Expiration:   expiration,
		hash:         nil,
		PublishSigns: []*crypto.Signature{},
		Delay:        delay,
		AmountLimit:  []*contract.Amount{},
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
	return common.Sha3(t.ToBytes(Base))
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
	return common.Sha3(t.ToBytes(Publish))
}

// ToPb convert tx to txpb.Tx for transmission.
func (t *Tx) ToPb() *txpb.Tx {
	tr := &txpb.Tx{
		Time:        t.Time,
		Expiration:  t.Expiration,
		GasLimit:    t.GasLimit,
		GasRatio:    t.GasRatio,
		Signers:     t.Signers,
		Delay:       t.Delay,
		ReferredTx:  t.ReferredTx,
		AmountLimit: t.AmountLimit,
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
	b, err := proto.Marshal(t.ToPb())
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
	t.GasRatio = tr.GasRatio
	t.Actions = []*Action{}
	t.Delay = tr.Delay
	t.ReferredTx = tr.ReferredTx
	t.AmountLimit = tr.AmountLimit
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
	err := proto.Unmarshal(b, tr)
	if err != nil {
		return err
	}
	t.FromPb(tr)
	return nil
}

// String return human-readable tx
func (t *Tx) String() string {
	if t == nil {
		return "<nil *tx.Tx>"
	}
	str := "Tx{\n"
	str += "	Time: " + strconv.FormatInt(t.Time, 10) + ",\n"
	str += "	Publisher: " + t.Publisher + ",\n"
	str += "	Action:\n"
	for _, a := range t.Actions {
		str += "		" + a.String()
	}
	str += "    AmountLimit:\n"
	str += fmt.Sprintf("%v", t.AmountLimit) + ",\n"
	str += "}\n"
	return str
}

// Hash return cached hash if exists, or calculate with Sha3.
func (t *Tx) Hash() []byte {
	if t.hash == nil {
		t.hash = common.Sha3(t.ToBytes(Full))
	}
	return t.hash
}

// IsDefer returns whether the transaction is a defer tx.
//
// Defer transaction is the transaction that's generated by a delay tx.
func (t *Tx) IsDefer() bool {
	return len(t.ReferredTx) > 0
}

// CanceledDelaytxHash returns the delay transaction hash that is canceled.
func (t *Tx) CanceledDelaytxHash() ([]byte, bool) {
	for _, action := range t.Actions {
		if action.Contract == "system.iost" && action.ActionName == "CancelDelaytx" {
			var actionData []string
			err := json.Unmarshal([]byte(action.Data), &actionData)
			if err == nil && len(actionData) > 0 {
				return common.Base58Decode(actionData[0]), true
			}
		}
	}
	return nil, false
}

func (t *Tx) verifyDeferBaseFields(referredTx *Tx) error {
	if referredTx.Time+referredTx.Delay != t.Time {
		return errors.New("unmatched referred tx delay time")
	}
	if referredTx.Expiration+referredTx.Delay != t.Expiration {
		return errors.New("unmatched referred tx expiration time")
	}
	if referredTx.GasRatio != t.GasRatio {
		return errors.New("unmatched referred tx gas ratio")
	}
	if referredTx.GasLimit != t.GasLimit {
		return errors.New("unmatched referred tx gas limit")
	}
	if len(referredTx.Actions) != len(t.Actions) {
		return errors.New("unmatched referred tx action length")
	}
	for i := 0; i < len(referredTx.Actions); i++ {
		if !referredTx.Actions[i].Equal(t.Actions[i]) {
			return errors.New("unmatched referred tx action")
		}
	}
	if len(referredTx.AmountLimit) != len(t.AmountLimit) {
		return errors.New("unmatched referred tx amount limit length")
	}
	for i := 0; i < len(referredTx.AmountLimit); i++ {
		if !referredTx.AmountLimit[i].Equal(t.AmountLimit[i]) {
			return errors.New("unmatched referred tx amount limit")
		}
	}
	return nil
}

func (t *Tx) verifyDeferSigFields(referredTx *Tx) error {
	if referredTx.Publisher != t.Publisher {
		return errors.New("unmatched referred tx publisher")
	}
	if len(referredTx.PublishSigns) != len(t.PublishSigns) {
		return errors.New("unmatched referred tx publishsigns length")
	}
	for i := 0; i < len(referredTx.PublishSigns); i++ {
		if !referredTx.PublishSigns[i].Equal(t.PublishSigns[i]) {
			return errors.New("unmatched referred tx publishsign")
		}
	}
	if len(referredTx.Signers) != len(t.Signers) {
		return errors.New("unmatched referred tx signers length")
	}
	for i := 0; i < len(referredTx.Signers); i++ {
		if referredTx.Signers[i] != t.Signers[i] {
			return errors.New("unmatched referred tx signer")
		}
	}
	if len(referredTx.Signs) != len(t.Signs) {
		return errors.New("unmatched referred tx signs length")
	}
	for i := 0; i < len(referredTx.Signs); i++ {
		if !referredTx.Signs[i].Equal(t.Signs[i]) {
			return errors.New("unmatched referred tx sign")
		}
	}
	return nil
}

// VerifyDefer verifes whether the defer tx is matched  with the referred tx.
func (t *Tx) VerifyDefer(referredTx *Tx) error {
	if err := t.verifyDeferBaseFields(referredTx); err != nil {
		return err
	}
	return t.verifyDeferSigFields(referredTx)
}

// VerifySelf verify tx's signature
func (t *Tx) VerifySelf() error { // only check whether sigs are legal
	if t.Delay > 0 && t.IsDefer() {
		return errors.New("invalid tx. including both delay and referredtx field")
	}
	// Defer tx does not need to verify signature.
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

// IsExpired checks whether the transaction is expired compared to the given time ct.
func (t *Tx) IsExpired(ct int64) bool {
	if t.Expiration <= ct {
		return true
	}
	if ct-t.Time > MaxExpiration {
		return true
	}
	return false
}

// IsCreatedBefore checks whether the transaction time is valid compared to the given time ct.
// ct may be time.Now().UnixNano() or block head time.
func (t *Tx) IsCreatedBefore(ct int64) bool {
	return t.Time <= ct
}

// CheckSize checks whether tx size is valid.
func (t *Tx) CheckSize() error {
	if len(t.ToBytes(Full)) > txSizeLimit {
		return fmt.Errorf("tx size illegal, should <= %v", txSizeLimit)
	}
	return nil
}

// CheckGas checks whether the transaction's gas is valid.
func (t *Tx) CheckGas() error {
	ratio := 100
	if t.GasRatio < minGasRatio || t.GasRatio > maxGasRatio {
		return fmt.Errorf("gas ratio illegal, should in [%v, %v]", minGasRatio/ratio, maxGasRatio/ratio)
	}
	if t.GasLimit < minGasLimit {
		return fmt.Errorf("gas limit illegal, should >= %v", minGasLimit/ratio)
	}
	if t.GasLimit > maxGasLimit {
		return fmt.Errorf("gas limit illegal, should <= %v", maxGasLimit/ratio)
	}
	return nil
}

// ToBytes converts tx to bytes.
func (t *Tx) ToBytes(l ToBytesLevel) []byte {
	sn := common.NewSimpleNotation()
	sn.WriteInt64(t.Time, true)
	sn.WriteInt64(t.Expiration, true)
	sn.WriteInt64(t.GasRatio, true)
	sn.WriteInt64(t.GasLimit, true)
	sn.WriteInt64(t.Delay, true)
	sn.WriteStringSlice(t.Signers, true)

	actionBytes := make([][]byte, 0, len(t.Actions))
	for _, a := range t.Actions {
		actionBytes = append(actionBytes, a.ToBytes())
	}
	sn.WriteBytesSlice(actionBytes, false)

	amountBytes := make([][]byte, 0, len(t.AmountLimit))
	for _, a := range t.AmountLimit {
		amountBytes = append(amountBytes, a.ToBytes())
	}
	sn.WriteBytesSlice(amountBytes, false)

	if l > Base {
		signBytes := make([][]byte, 0, len(t.Signs))
		for _, sig := range t.Signs {
			signBytes = append(signBytes, sig.ToBytes())
		}
		sn.WriteBytesSlice(signBytes, false)
	}

	if l > Publish {
		sn.WriteBytes(t.ReferredTx, true)
		sn.WriteString(t.Publisher, true)

		signBytes := make([][]byte, 0, len(t.PublishSigns))
		for _, sig := range t.PublishSigns {
			signBytes = append(signBytes, sig.ToBytes())
		}
		sn.WriteBytesSlice(signBytes, false)
	}

	return sn.Bytes()
}
