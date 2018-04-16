package account

import (
  "fmt"
  "github.com/iost-official/prototype/all_hash"
  "github.com/LoCCS/bliss"
  "github.com/LoCCS/bliss/sampler"
  "github.com/LoCCS/bliss/params"
)

const (
  AddressLength = 52  // or 43 if use base64
  BlissVersion = params.BLISS_B_4
)

type Address [AddressLength]byte

func (addr Address) ToString() string {return string(addr[:])}

func ByteToAddress(bs []byte) Address {
  b_len := len(bs)
  var addr Address

  if b_len < AddressLength {
    for i, b := range bs {
      addr[AddressLength - b_len + i] = b
    }
  } else {
    for i, _ := range addr {
      addr[i] = bs[i]
    }
  }

  return addr
}

func StringToAddress(str string) Address {return ByteToAddress(([]byte)(str))}

func GenerateAddress(passphrase string) (Address, error) {
  raw_len := uint32(len(passphrase))
  var seed_size uint32
  if raw_len < sampler.SHA_512_DIGEST_LENGTH {
    seed_size = sampler.SHA_512_DIGEST_LENGTH
  } else {
    seed_size = raw_len
  }

  seed := make([]uint8, seed_size)
  copy(seed[:raw_len], ([]uint8)(passphrase))

  entropy, err := sampler.NewEntropy(seed)
  if err != nil {
    var tmp_addr Address
    return tmp_addr, fmt.Errorf("Bad passphrase.")
  }

  sk, err := bliss.GeneratePrivateKey(BlissVersion, entropy)
  if err != nil {
    var tmp_addr Address
    return tmp_addr, fmt.Errorf("Bad passphrase.")
  }

  pk := sk.PublicKey()
  return StringToAddress((all_hash.Sha3_256(pk.Encode())).ToBase32Hex()), nil   // or ToBase64URL
}
