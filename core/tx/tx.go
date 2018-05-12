package tx

import (
	"fmt"
	"time"

	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
)

//go:generate gencode go -schema=structs.schema -package=tx

// Transaction 的实现
type Tx struct {
	Time      int64
	Nonce     int64
	Contract  vm.Contract
	Signs     []common.Signature
	Publisher common.Signature
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
func SignContract(tx Tx, account account.Account) (Tx, error) {
	sign, err := common.Sign(common.Secp256k1, tx.baseHash(), account.Seckey)
	if err != nil {
		return tx, err
	}
	tx.Signs = append(tx.Signs, sign)
	return tx, nil
}

// 合约的发布者进行签名，此签名的用户用于支付gas
func SignTx(tx Tx, account account.Account) (Tx, error) {
	sign, err := common.Sign(common.Secp256k1, tx.publishHash(), account.Seckey)
	if err != nil {
		return tx, err
	}
	tx.Publisher = sign
	return tx, nil
}

func (t *Tx) baseHash() []byte {
	tbr := TxBaseRaw{t.Time, t.Nonce, t.Contract.Encode()}
	b, err := tbr.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return common.Sha256(b)
}

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

func (t *Tx) Encode() []byte {
	s := make([][]byte, 0)
	for _, sign := range t.Signs {
		s = append(s, sign.Encode())
	}
	tr := TxRaw{t.Time, t.Nonce, t.Contract.Encode(), s, t.Publisher.Encode()}
	b, err := tr.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return b
}
func (t *Tx) Decode(b []byte) error {
	var tr TxRaw
	_, err := tr.Unmarshal(b)
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
	t.Contract.SetPrefix(vm.HashToPrefix(t.Hash()))
	for _, sign := range t.Signs {
		t.Contract.AddSigner(vm.PubkeyToIOSTAccount(sign.Pubkey))
	}
	t.Nonce = tr.Nonce
	t.Time = tr.Time
	return nil
}
func (t *Tx) Hash() []byte {
	return common.Sha256(t.Encode())
}

// 验证签名的函数
func (t *Tx) VerifySelf() error {
	baseHash := t.baseHash()
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
