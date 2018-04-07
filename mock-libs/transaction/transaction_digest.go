package transaction

import (
	"crypto"
)

func (this *TransactionType) merkle_digest() DigestType {
	var enc EncoderType
	crypto.pack(enc, *this)
	return enc.result()
}

func (this *TransactionType) digest() DigestType {
	var enc EncoderType
	crypto.pack(enc, this.chain_id)
	crypto.pack(enc, *this)
	return enc.result()
}
