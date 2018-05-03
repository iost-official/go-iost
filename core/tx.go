package core

import (
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/vm"
)

type Tx struct {
	Time      int64
	Nonce     int64
	Contract  vm.Contract
	Signs     []common.Signature
	Publisher common.Signature
}

func (t *Tx) BaseHash() []byte {
	tbr := TxBaseRaw{t.Time, t.Nonce, t.Contract.Encode()}
	b, err := tbr.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return common.Sha256(b)
}

func (t *Tx) PublishHash() []byte {
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
	err = t.Contract.Decode(tr.Contract)
	if err != nil {
		return err
	}
	t.Contract.SetSender(t.Publisher.Pubkey)
	t.Contract.SetPrefix(string(t.Hash()))
	for _, sign := range t.Signs {
		t.Contract.AddSigner(sign.Pubkey)
	}
	t.Nonce = tr.Nonce
	t.Time = tr.Time
	return nil
}
func (t *Tx) Hash() []byte {
	return common.Sha256(t.Encode())
}
