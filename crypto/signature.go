package crypto

import (
	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/common"
)

// Signature is the signature of some message
type Signature struct {
	Algorithm Algorithm

	Sig    []byte
	Pubkey []byte
}

// NewSignature returns new signature
func NewSignature(algo Algorithm, info []byte, privkey []byte) *Signature {
	s := &Signature{
		Algorithm: algo,
		Sig:       algo.Sign(info, privkey),
		Pubkey:    algo.GetPubkey(privkey),
	}
	return s
}

// Verify will verify the info
func (s *Signature) Verify(info []byte) bool {
	return s.Algorithm.Verify(info, s.Pubkey, s.Sig)
}

// SetPubkey will set the public key
func (s *Signature) SetPubkey(pubkey []byte) {
	s.Pubkey = pubkey
}

// Encode will marshal the signature by protobuf
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

// Decode will unmarshal the signature by protobuf
func (s *Signature) Decode(b []byte) error {
	sr := &SignatureRaw{}
	err := proto.Unmarshal(b, sr)
	if err != nil {
		return err
	}
	s.Algorithm = Algorithm(sr.Algorithm)
	s.Sig = sr.Sig
	s.Pubkey = sr.PubKey
	return err
}

// Hash returns the hash code of signature
func (s *Signature) Hash() []byte {
	b, _ := s.Encode()
	return common.Sha3(b)
}
