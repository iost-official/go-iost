package transaction

import (
	"crypto"
)

func (this *SignatureType) s_sign(key PrivateKeyType, chain_id ChainIdType) SignatureType {
	h := this.sig_digest(chain_id)
	this.signatures.append(this.key.sign_compact(h))
	return this.signatures[len(this.signatures)-1]
}

func (this *SignatureType) s_sign(key PrivateKeyType, chain_id ChainIdType) SignatureType {
	var enc encoder
	crypto.pack(enc, chain_id)
	crypto.pack(enc, this)
	return this.key.sign_compact(enc.result())
}

func (this *TransactionType) set_Expiration(time TimeType) {
	this.expiration = time
}

func (this *TransactionType) set_refer_block(reference BlockIdType) {
	this.ref_block_num = crypto.reverse(reference.hash[0])
	this.ref_block_prefix = reference.hash[0]
}
