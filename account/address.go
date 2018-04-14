package account

import (
  //"github.com/iost-official/prototype/all_hash"
  "github.com/LoCCS/bliss"
)

const (
  UserAddressLength = 52  // or 43 if use base64
)

type UserAddress [UserAddressLength]byte

func GenerateUserAddress(pk bliss.PublicKey) UserAddress {
  // todo
  var tmp UserAddress
  return tmp
}
