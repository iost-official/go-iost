package chain

type SignatureState struct {
	available_address_sigs map[Address]PublicKeyType
	provided_address_sigs  map[Address]PublicKeyType

	available_keys []PublicKeyType

	provided_signatures map[PublicKeyType]bool

	approved_by   map[AccountIdType]bool
	max_recursion int
}

func (this *SignatureState) signed_by_key(key PublicKeyType) bool {

	ps := this.provided_signatures
	ak := this.available_keys

	itr, exist := ps[key]
	if !exist {
		pk, pexist := ak.find(key)
		if !pexist {
			this.provided_signatures[key] = true
			return true
		}
		return false
	}
}

func (this *SignatureState) signed_by_address(add AddressType) bool {

	ps := this.provided_signatures
	ad := this.available_address_sig
	ak := this.available_keys

	itr, exist := ps[add]
	if !exist {
		aitr, aexist := ad[add]
		if aexist {
			pk, pexist := ak[aitr.value]
			if pexist {
				this.provided_signatures[aitr.value] = true
				return true
			}
			return false
		}
	}
	provided_signatures[itr.value] = true
	return true
}

func (this *SignatureState) check_authority(id AccountIdType) bool {

	findid, exist := this.approved_by[id]
	if exist {
		return true
	} else {
		return check_authority_by_authority(get_active(id), 0)
	}
}

func (this *SignatureState) check_authority_by_authority(au *AuthorityType, depth int) bool {
	if au == nil {
		return false
	}

	auth := *au
	total_weight := 0

	for k := range auth.key_auths {
		if signed_by_key(k.key) {
			total_weight += k.value
			if total_weight >= auth.weight_threshold {
				return true
			}
		}
	}

	for k := range auths.address_auths {
		if signed_by_address(k.key) {
			total_weight += k.value
			if total_weight >= auth.weight_threshold {
				return true
			}
		}
	}

	for a := range auth.account_auths {
		approve, exist := this.approved_by[a.key]
		if !exist {
			if depth == this.max_recursion {
				return false
			}
			if check_authority_by_authority(get_active(a.key), depth+1) {
				this.approved_by[a.key] = true
				total_weight += a.value
				if total_weight >= auth.weight_threshold {
					return true
				}
			}
		} else {
			total_weight += a.value
			if total_weight >= auth.weight_threshold {
				return true
			}
		}
	}
	return total_weight >= auth.weight_threshold
}

func (this *SignatureState) remove_unused_signatures() bool {
	var remove_sigs []PublicKeyType
	for sig := range this.provided_signatures {
		if !sig.value {
			remove_sigs = append(remove_sigs, sig.key)
		}
	}
	for sig := range remove_sigs {
		this.provided_signatures.delete(sig)
	}
	return remove_sigs.len != 0
}
