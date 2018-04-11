package bliss

import (
	"fmt"
	"github.com/LoCCS/bliss/huffman"
	"github.com/LoCCS/bliss/params"
	"github.com/LoCCS/bliss/poly"
	"github.com/LoCCS/bliss/sampler"
)

type PrivateKey struct {
	s1 *poly.PolyArray
	s2 *poly.PolyArray
	a  *poly.PolyArray
}

type PublicKey struct {
	a *poly.PolyArray
}

func GeneratePrivateKey(version int, entropy *sampler.Entropy) (*PrivateKey, error) {
	// Generate g
	s2 := poly.UniformPoly(version, entropy)
	if s2 == nil {
		return nil, fmt.Errorf("Failed to generate uniform polynomial g")
	}
	// s2 = 2g-1
	s2.ScalarMul(2)
	s2.ScalarInc(-1)

	t, err := s2.NTT()
	if err != nil {
		return nil, err
	}

	for j := 0; j < 4; j++ {
		s1 := poly.UniformPoly(version, entropy)
		if s1 == nil {
			return nil, fmt.Errorf("Failed to generate uniform polynomial f")
		}
		u, err := s1.NTT()
		if err != nil {
			return nil, err
		}
		u, err = u.InvertAsNTT()
		if err != nil {
			continue
		}
		t.MulModQ(u)
		t, err = t.INTT()
		if err != nil {
			return nil, err
		}
		t.ScalarMulModQ(-1)
		a, err := t.NTT()
		if err != nil {
			return nil, err
		}
		key := PrivateKey{s1, s2, a}
		return &key, nil
	}
	return nil, fmt.Errorf("Failed to generate invertible polynomial")
}

func (privateKey *PrivateKey) PublicKey() *PublicKey {
	return &PublicKey{privateKey.a}
}

func (privateKey *PrivateKey) Param() *params.BlissBParam {
	return privateKey.s1.Param()
}

func (privateKey *PrivateKey) Destroy() {
	n := privateKey.Param().N
	s1data := privateKey.s1.GetData()
	s2data := privateKey.s2.GetData()
	for i := 0; i < int(n); i++ {
		s1data[i] = 0
		s2data[i] = 0
	}
}

func (publicKey *PublicKey) Param() *params.BlissBParam {
	return publicKey.a.Param()
}

func (privateKey *PrivateKey) String() string {
	return fmt.Sprintf("{s1:%s,s2:%s,a:%s}",
		privateKey.s1.String(), privateKey.s2.String(), privateKey.a.String())
}

func (publicKey *PublicKey) String() string {
	return fmt.Sprintf("{a:%s}", publicKey.a.String())
}

func (publicKey *PublicKey) Encode() []byte {
	n := publicKey.Param().N
	data := publicKey.a.GetData()
	ret := make([]byte, n*2+1)
	ret[0] = byte(publicKey.Param().Version)
	for i := 0; i < int(n); i++ {
		ret[i*2+1] = byte(uint16(data[i]) >> 8)
		ret[i*2+2] = byte(uint16(data[i]) & 0xff)
	}
	return ret[:]
}

func DecodePublicKey(data []byte) (*PublicKey, error) {
	a, err := poly.New(int(data[0]))
	if err != nil {
		return nil, fmt.Errorf("Error in generating new polyarray: %s", err.Error())
	}
	ret := &PublicKey{a}
	n := a.Param().N
	retdata := a.GetData()
	for i := 0; i < int(n); i++ {
		retdata[i] = (int32(data[i*2+1]) << 8) | (int32(data[i*2+2]))
	}
	return ret, nil
}

func (privateKey *PrivateKey) Encode() []byte {
	n := privateKey.Param().N
	s1data := privateKey.s1.GetData()
	s2data := privateKey.s2.GetData()
	ret := make([]byte, n*2+1)
	ret[0] = byte(privateKey.Param().Version)
	s1 := ret[1 : 1+n]
	s2 := ret[1+n:]
	for i := 0; i < int(n); i++ {
		s1[i] = byte(s1data[i] + 4)
		s2[i] = byte(s2data[i] + 4)
	}
	return ret[:]
}

func DecodePrivateKey(data []byte) (*PrivateKey, error) {
	s1, err := poly.New(int(data[0]))
	if err != nil {
		return nil, fmt.Errorf("Error in generating new polyarray: %s", err.Error())
	}
	s2, err := poly.NewPolyArray(s1.Param())
	if err != nil {
		return nil, fmt.Errorf("Error in generating new polyarray: %s", err.Error())
	}

	// Recover f,g from the bytes
	// then everything is like the key generation procedure
	n := s1.Param().N
	s1data := s1.GetData()
	s2data := s2.GetData()
	s1src := data[1 : 1+n]
	s2src := data[1+n:]
	for i := 0; i < int(n); i++ {
		s1data[i] = int32(s1src[i]) - 4
		s2data[i] = int32(s2src[i]) - 4
	}

	t, err := s2.NTT()
	if err != nil {
		return nil, err
	}
	u, err := s1.NTT()
	if err != nil {
		return nil, err
	}
	u, err = u.InvertAsNTT()
	if err != nil {
		return nil, err
	}
	t.MulModQ(u)
	t, err = t.INTT()
	if err != nil {
		return nil, err
	}
	t.ScalarMulModQ(-1)
	a, err := t.NTT()
	if err != nil {
		return nil, err
	}
	key := PrivateKey{s1, s2, a}
	return &key, nil
}

func (privateKey *PrivateKey) Serialize() []byte {
	packer := huffman.NewBitPacker()
	n := privateKey.Param().N
	s1data := privateKey.s1.GetData()
	s2data := privateKey.s2.GetData()
	for i := 0; i < int(n); i++ {
		packer.WriteBits(uint64(s1data[i]+2), 3)
	}
	packer.WriteBits(uint64((s2data[0]+1)/2)+2, 3)
	for i := 1; i < int(n); i++ {
		packer.WriteBits(uint64(s2data[i]/2+2), 3)
	}
	ret := []byte{byte(privateKey.Param().Version)}
	return append(ret, packer.Data()...)
}

func DeserializePrivateKey(data []byte) (*PrivateKey, error) {
	s1, err := poly.New(int(data[0]))
	if err != nil {
		return nil, fmt.Errorf("Error in generating new polyarray: %s", err.Error())
	}
	s2, err := poly.NewPolyArray(s1.Param())
	if err != nil {
		return nil, fmt.Errorf("Error in generating new polyarray: %s", err.Error())
	}

	n := s1.Param().N
	unpacker := huffman.NewBitUnpacker(data[1:], 6*n)
	s1data := s1.GetData()
	s2data := s2.GetData()
	for i := 0; i < int(n); i++ {
		bits, err := unpacker.ReadBits(3)
		if err != nil {
			return nil, err
		}
		s1data[i] = int32(bits) - 2
	}
	bits, err := unpacker.ReadBits(3)
	if err != nil {
		return nil, err
	}
	s2data[0] = (int32(bits)-2)*2 - 1
	for i := 1; i < int(n); i++ {
		bits, err := unpacker.ReadBits(3)
		if err != nil {
			return nil, err
		}
		s2data[i] = (int32(bits) - 2) * 2
	}
	t, err := s2.NTT()
	if err != nil {
		return nil, err
	}
	u, err := s1.NTT()
	if err != nil {
		return nil, err
	}
	u, err = u.InvertAsNTT()
	if err != nil {
		return nil, err
	}
	t.MulModQ(u)
	t, err = t.INTT()
	if err != nil {
		return nil, err
	}
	t.ScalarMulModQ(-1)
	a, err := t.NTT()
	if err != nil {
		return nil, err
	}
	key := PrivateKey{s1, s2, a}
	return &key, nil
}

func (publicKey *PublicKey) Serialize() []byte {
	qbit := publicKey.Param().Qbits
	n := publicKey.Param().N
	packer := huffman.NewBitPacker()
	adata := publicKey.a.GetData()
	for i := 0; i < int(n); i++ {
		packer.WriteBits(uint64(adata[i]), qbit)
	}
	ret := []byte{byte(publicKey.Param().Version)}
	return append(ret, packer.Data()...)
}

func DeserializePublicKey(data []byte) (*PublicKey, error) {
	a, err := poly.New(int(data[0]))
	if err != nil {
		return nil, fmt.Errorf("Error in generating new polyarray: %s", err.Error())
	}
	n := a.Param().N
	qbit := a.Param().Qbits
	unpacker := huffman.NewBitUnpacker(data[1:], n*qbit)
	adata := a.GetData()
	for i := 0; i < int(n); i++ {
		bits, err := unpacker.ReadBits(qbit)
		if err != nil {
			return nil, err
		}
		adata[i] = int32(bits)
	}
	return &PublicKey{a}, nil
}
