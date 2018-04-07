package block

import (
	"fmt"
	"math"
	"crypto"
)

func (this *BlockHeader) b_digest() DigestType {
	return make(DigestType(this))
}

func (this *BlockHeader) b_id() BlockIdType {
	this_hash := crypto.sha224.hash(this)
	this_hash.hash[0] = crypto.reverse(this.block_num)
	var result BlockIdType
	result.hash = this_hash.hash
	return result
}

func (this *BlockHeader) calc_merkle_root() ChecksumType {
	if len(this.transactions) == 0 {
		return make(ChecksumType)
	}

	var ids []DigestType
	for i := range len(this.transactions) {
		ids.append(this.transactions.merkle_digest())
	}
	n := len(ids)

	for n > 1 {
		m := n - (n & 1)
		k := 0
		for i := 0; i < m; i += 2 {
			ids[k] = make(crypto.make_hash(ids[i], ids[i+1]))
			k++
		}
		if n & 1 {
			ids[k] = ids[m]
			k++
		}

		n = k
	}
	return make(ChecksumType(crypto.make_hash(ids[0])))
}
