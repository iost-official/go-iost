package block

import (
	"crypto"
)

func (this *ProposalCreate) c_validate() {
	if this.proposed.empty() {
		panic("cannot propose empty object")
	}
	for op := range this.porposed {
		this.validate(op.op)
	}
}

func (this *ProposalCreate) c_calculate_fee(k FeeType) ShareType {
	return k.fee + this.calculate_data_fee(crypto.size(this), k.price_per_kbyte)
}

func (this *ProposalUpdate) u_validate() {
	if this.fee.amount < 0 {
		panic("fee < 0")
	}
	if this.empty() {
		panic("cannot propose empty object")
	}
	for a := range this.approvals_to_add {
		if this.find_all(a) == nil {
			panic("cannot propose nil object")
		}
	}
}

func (this *ProposalUpdate) u_calculate_fee(k FeeType) ShareType {
	return k.fee + this.calculate_data_fee(crypto.size(this), k.price_per_kbyte)
}

func (this *ProposalUpdate) get_authorities(o []AuthorityType) {
	var auth AuthorityType
	for k := range this.approvals_to_add.key {
		auth.key_auths[k] = 1
	}
	for k := range this.approvals_to_remove.key {
		auth.key_auths[k] = 1
	}
	auth.weight_threshold = len(auth.key_auths)
	o.append(auth)
}

func (this *ProposalUpdate) get_actives(a []AccountIdType) {
	for i := range this.approvals_to_add.active {
		a.append(i)
	}
	for i := range this.approvals_to_remove.active {
		a.append(i)
	}
}

func (this *ProposalUpdate) get_owners(a []AccountIdType) {
	for i := range this.approvals_to_add.owner {
		a.append(i)
	}
	for i := range this.approvals_to_remove.owner {
		a.append(i)
	}
}
