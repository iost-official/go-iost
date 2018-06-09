package tx

import (
	"fmt"
	"time"

	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	"strconv"
)

//go:generate gencode go -schema=structs.schema -package=tx

// Transaction 的实现
type Tx struct {
	Time      int64
	Nonce     int64
	Contract  vm.Contract
	Signs     []common.Signature
	Publisher common.Signature
	Recorder  common.Signature
}

// 新建一个Tx，需要通过编译器得到一个contract
func NewTx(nonce int64, contract vm.Contract) Tx {
	return Tx{
		Time:     time.Now().UnixNano(),
		Nonce:    nonce,
		Contract: contract,
	}
}

// 合约的参与者进行签名
func SignContract(tx Tx, account account.Account) (common.Signature, error) {
	sign, err := common.Sign(common.Secp256k1, tx.BaseHash(), account.Seckey)
	if err != nil {
		return sign, err
	}
	return sign, nil
}

// 合约的发布者进行签名，此签名的用户用于支付gas
func SignTx(tx Tx, account account.Account, signs ...common.Signature) (Tx, error) {
	tx.Signs = append(tx.Signs, signs...)
	sign, err := common.Sign(common.Secp256k1, tx.publishHash(), account.Seckey)
	if err != nil {
		return tx, err
	}
	tx.Publisher = sign
	return tx, nil
}

func RecordTx(tx Tx, account account.Account) (Tx, error) {
	sign, err := common.Sign(common.Secp256k1, tx.BaseHash(), account.Seckey)
	if err != nil {
		return tx, err
	}
	tx.Recorder = sign
	return tx, nil
}

// Time,Noce,Contract形成的基本哈希值
func (t *Tx) BaseHash() []byte {
	tbr := TxBaseRaw{t.Time, t.Nonce, t.Contract.Encode()}
	b, err := tbr.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return common.Sha256(b)
}

// 发布者使用的hash值，包含参与者的签名
func (t *Tx) publishHash() []byte {
	s := make([][]byte, 0)
	for _, sign := range t.Signs {
		s = append(s, sign.Encode())
	}
	tpr := TxPublishRaw{t.Time, t.Nonce, t.Contract.Encode(), s}
	b, err := tpr.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return common.Sha256(b)
}

// 对Tx进行编码
func (t *Tx) Encode() []byte {
	s := make([][]byte, 0)
	for _, sign := range t.Signs {
		s = append(s, sign.Encode())
	}
	tr := TxRaw{t.Time, t.Nonce, t.Contract.Encode(), s, t.Publisher.Encode(), []byte{}}
	b, err := tr.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return b
}

// 对Tx进行解码
func (t *Tx) Decode(b []byte) (err error) {
	var tr TxRaw
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	_, err = tr.Unmarshal(b)
	if err != nil {
		return err
	}
	t.Publisher.Decode(tr.Publisher)
	for _, sr := range tr.Signs {
		var sign common.Signature
		err = sign.Decode(sr)
		if err != nil {
			return err
		}
		t.Signs = append(t.Signs, sign)
	}
	if t.Contract == nil {
		switch tr.Contract[0] {
		case 0:
			t.Contract = &lua.Contract{}
			t.Contract.Decode(tr.Contract)

		default:
			return fmt.Errorf("Tx.Decode:tx.Contract syntax error")
		}
	} else {
		err = t.Contract.Decode(tr.Contract)
	}

	if err != nil {
		return err
	}
	t.Contract.SetSender(vm.PubkeyToIOSTAccount(t.Publisher.Pubkey))
	for _, sign := range t.Signs {
		t.Contract.AddSigner(vm.PubkeyToIOSTAccount(sign.Pubkey))
	}
	t.Nonce = tr.Nonce
	t.Time = tr.Time
	t.Contract.SetPrefix(vm.HashToPrefix(t.Hash()))
	return nil
}

// 计算Tx的哈希值
func (t *Tx) Hash() []byte {
	return common.Sha256(t.Encode())
}

// 公钥+nonoc 可用于交易判重
func (t *Tx) TxID() string {
	hash := string(t.Publisher.Pubkey)+strconv.FormatInt(t.Nonce,10)+strconv.FormatInt(t.Time,10)
	return hash
}

// 验证签名的函数
func (t *Tx) VerifySelf() error {
	baseHash := t.BaseHash() // todo 在basehash内缓存，不需要在应用进行缓存
	for _, sign := range t.Signs {
		ok := common.VerifySignature(baseHash, sign)
		if !ok {
			return fmt.Errorf("signer error")
		}
	}

	ok := common.VerifySignature(t.publishHash(), t.Publisher)
	if !ok {
		return fmt.Errorf("publisher error")
	}
	return nil
}

func (t *Tx) VerifySigner(sig common.Signature) bool {
	return common.VerifySignature(t.BaseHash(), sig)
}

type TransactionsList []*Tx

func (s TransactionsList) Len() int           { return len(s) }
func (s TransactionsList) Less(i, j int) bool { return s[i].Contract.Info().Price < s[j].Contract.Info().Price }
func (s TransactionsList) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s *TransactionsList) Push(x interface{}) {
	*s = append(*s, x.(*Tx))
}

func (s *TransactionsList) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

func DiffTxList(a, b TransactionsList) (keep TransactionsList) {
	keep = make(TransactionsList, 0, len(a))

	remove := make(map[string]struct{})
	for _, tx := range b {
		remove[string(tx.Hash())] = struct{}{}
	}

	for _, tx := range a {
		if _, ok := remove[string(tx.Hash())]; !ok {
			keep = append(keep, tx)
		}
	}

	return keep
}
