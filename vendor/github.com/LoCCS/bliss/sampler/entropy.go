package sampler

import (
	"fmt"
	"golang.org/x/crypto/sha3"
)

const (
	SHA_512_DIGEST_LENGTH uint32 = 64
	EPOOL_HASH_COUNT             = 10
	CHAR_POOL_SIZE               = SHA_512_DIGEST_LENGTH * EPOOL_HASH_COUNT
	INT16_POOL_SIZE              = SHA_512_DIGEST_LENGTH / 2 * EPOOL_HASH_COUNT
	INT64_POOL_SIZE              = SHA_512_DIGEST_LENGTH / 8 * EPOOL_HASH_COUNT
)

type Entropy struct {
	bitpool   uint64
	charpool  []uint8
	int16pool []uint16
	int64pool []uint64
	seed      []uint8

	bitp   uint32
	charp  uint32
	int16p uint32
	int64p uint32
}

func NewEntropy(seed []uint8) (*Entropy, error) {
	if len(seed) < int(SHA_512_DIGEST_LENGTH) {
		return nil, fmt.Errorf("Insufficient seed length, need %d, got %d",
			SHA_512_DIGEST_LENGTH, len(seed))
	}
	entropy := Entropy{0, []uint8{}, []uint16{}, []uint64{}, []uint8{}, 0, 0, 0, 0}
	entropy.charpool = make([]uint8, CHAR_POOL_SIZE)
	entropy.int16pool = make([]uint16, INT16_POOL_SIZE)
	entropy.int64pool = make([]uint64, INT64_POOL_SIZE)
	entropy.seed = make([]uint8, SHA_512_DIGEST_LENGTH)
	for i := 0; i < int(SHA_512_DIGEST_LENGTH); i++ {
		entropy.seed[i] = seed[i]
	}
	entropy.refreshCharPool()
	entropy.refreshInt16Pool()
	entropy.refreshInt64Pool()
	entropy.refreshBitPool()
	return &entropy, nil
}

func (entropy *Entropy) incrementSeed() {
	for i := 0; i < int(SHA_512_DIGEST_LENGTH); i++ {
		entropy.seed[i]++
		if entropy.seed[i] > 0 {
			break
		}
	}
}

func (entropy *Entropy) refreshCharPool() {
	for i := 0; i < int(EPOOL_HASH_COUNT); i++ {
		offset := i * int(SHA_512_DIGEST_LENGTH)
		sha := sha3.Sum512([]byte(entropy.seed))
		for j := 0; j < int(SHA_512_DIGEST_LENGTH); j++ {
			entropy.charpool[offset+j] = uint8(sha[j])
		}
		entropy.incrementSeed()
	}
	entropy.charp = 0
}

func (entropy *Entropy) refreshInt16Pool() {
	for i := 0; i < int(EPOOL_HASH_COUNT); i++ {
		offset := i * int(SHA_512_DIGEST_LENGTH) / 2
		sha := sha3.Sum512([]byte(entropy.seed))
		for j := 0; j < int(SHA_512_DIGEST_LENGTH)/2; j++ {
			entropy.int16pool[offset+j] = combineUint16(sha[:], j*2)
		}
		entropy.incrementSeed()
	}
	entropy.int16p = 0
}

func (entropy *Entropy) refreshInt64Pool() {
	for i := 0; i < int(EPOOL_HASH_COUNT); i++ {
		offset := i * int(SHA_512_DIGEST_LENGTH) / 8
		sha := sha3.Sum512([]byte(entropy.seed))
		for j := 0; j < int(SHA_512_DIGEST_LENGTH)/8; j++ {
			entropy.int64pool[offset+j] = combineUint64(sha[:], j*8)
		}
		entropy.incrementSeed()
	}
	entropy.int64p = 0
}

func (entropy *Entropy) refreshBitPool() {
	entropy.bitpool = entropy.Uint64()
	entropy.bitp = 0
}

func (entropy *Entropy) Uint64() uint64 {
	if entropy.int64p >= INT64_POOL_SIZE {
		entropy.refreshInt64Pool()
	}
	ret := entropy.int64pool[entropy.int64p]
	entropy.int64p++
	return ret
}

func (entropy *Entropy) Uint16() uint16 {
	if entropy.int16p >= INT16_POOL_SIZE {
		entropy.refreshInt16Pool()
	}
	ret := entropy.int16pool[entropy.int16p]
	entropy.int16p++
	return ret
}

func (entropy *Entropy) Char() uint8 {
	if entropy.charp >= CHAR_POOL_SIZE {
		entropy.refreshCharPool()
	}
	ret := entropy.charpool[entropy.charp]
	entropy.charp++
	return ret
}

func (entropy *Entropy) UintBit() uint64 {
	if entropy.bitp >= 64 {
		entropy.refreshBitPool()
	}
	bit := entropy.bitpool & 1
	entropy.bitpool >>= 1
	entropy.bitp++
	return bit
}

func (entropy *Entropy) Bit() bool {
	return entropy.UintBit() == 1
}

func (entropy *Entropy) Bits(n int) uint32 {
	ret := uint32(0)
	for n > 0 {
		ret <<= 1
		ret |= uint32(1 & entropy.UintBit())
		n--
	}
	return ret
}
