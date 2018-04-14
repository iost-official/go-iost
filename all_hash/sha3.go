package all_hash

import (
  "github.com/InWeCrypto/sha3"
)

func Sha3_256(raw []byte) Hash {return Hash(sha3.Sum256(raw))}

func Sha3_512(raw []byte) LongHash {return LongHash(sha3.Sum512(raw))}
