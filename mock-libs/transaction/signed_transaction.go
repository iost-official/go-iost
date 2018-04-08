package transaction

import (
	"github.com/iost-official/Go-IOS-Protocol/mock-libs/block"
)

func (this *SignedTransaction) get_signature_keys(chain_id ChainIdType) map[PublicKeyType]bool {
	d := this.sig_digest(chain_id)
	var result map[PublicKeyType]bool
	for sig := range this.signatures {
		result[make(PublicKeyType(sig, d))] = true
	}
	return result
}

func (this *SignedTransaction) get_required_signatures(chain_id ChainIdType, available_keys map[PublicKeyType]bool) map[PublicKeyType]bool {

	var required_active, required_owner map[AccountIdType]bool
	var other []authority
	this.get_required_authorities(required_active, required_owner, other)

	s := make(SignatureState(get_signature_keys(chain_id), get_active, available_keys))
	s.max_recursion = this.max_recursion_depth

	for auth := range other {
		s.check_authority_by_authority(&auth, 0)
	}

	for owenr := range required_owner {
		s.check_authority(get_owner(owner))
	}

	for active := range required_active {
		s.check_authority(active)
	}

	s.remove_unused_signatures()

	var result map[PublicKeyType]bool

	for provided_sig := range s.provided_signatures {
		ak, exist := this.available[provided_sig.key]
		if exist {
			result[provided_sig.key] = true
		}
	}
	return result
}
