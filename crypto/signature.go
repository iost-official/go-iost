package crypto

import (
	"errors"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto/pb"
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

// ToPb convert Signature to proto buf data structure.
func (s *Signature) ToPb() *sigpb.Signature {
	return &sigpb.Signature{
		Algorithm: int32(s.Algorithm),
		Sig:       s.Sig,
		PubKey:    s.Pubkey,
	}
}

// FromPb convert Signature from proto buf data structure.
func (s *Signature) FromPb(sr *sigpb.Signature) *Signature {
	s.Algorithm = Algorithm(sr.Algorithm)
	s.Sig = sr.Sig
	s.Pubkey = sr.PubKey
	return s
}

// Encode will marshal the signature by protobuf
func (s *Signature) Encode() ([]byte, error) {
	b, err := s.ToPb().Marshal()
	if err != nil {
		return nil, errors.New("fail to encode signature")
	}
	return b, nil
}

// Decode will unmarshal the signature by protobuf
func (s *Signature) Decode(b []byte) error {
	sr := &sigpb.Signature{}
	err := sr.Unmarshal(b)
	if err != nil {
		return err
	}
	s.FromPb(sr)
	return nil
}

// Hash returns the hash code of signature
func (s *Signature) Hash() []byte {
	b, _ := s.Encode()
	return common.Sha3(b)
}
