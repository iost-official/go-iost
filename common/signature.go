package common

import (
	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/Go-IOS-Protocol/crypto"
)

//go:generate gencode go -schema=structs.schema -package=verifier

type SignMode bool

const (
	SavePubkey SignMode = true
	NilPubkey  SignMode = false
)

type Signature struct {
	Algorithm crypto.Algorithm

	Sig    []byte
	Pubkey []byte
}

func Sign(algo crypto.Algorithm, info, privkey []byte, smode SignMode) Signature {
	s := Signature{Pubkey: nil}
	s.Algorithm = algo
	if smode {
		s.Pubkey = s.Algorithm.GetPubkey(privkey)
	}
	s.Sig = s.Algorithm.Sign(info, privkey)
	return s
}

func VerifySignature(info []byte, s Signature) bool {
	return s.Algorithm.Verify(info, s.Pubkey, s.Sig)
}

func (s *Signature) SetPubkey(pubkey []byte) {
	s.Pubkey = pubkey
}

func (s *Signature) Encode() ([]byte, error) {
	sr := &SignatureRaw{
		Algorithm: int32(s.Algorithm),
		Sig:       s.Sig,
		PubKey:    s.Pubkey,
	}
	b, err := proto.Marshal(sr)
	if err != nil {
		return nil, errors.New("fail to encode signature")
	}
	return b, nil
}

func (s *Signature) Decode(b []byte) error {
	sr := &SignatureRaw{}
	err := proto.Unmarshal(b, sr)
	if err != nil {
		return err
	}
	s.Algorithm = crypto.Algorithm(sr.Algorithm)
	s.Sig = sr.Sig
	s.Pubkey = sr.PubKey
	return err
}

func (s *Signature) Hash() []byte {
	b, _ := s.Encode()
	return Sha3(b)
}
