package transaction

import (
	"fmt"
	"math"

	"crypto"
	"iosbase/debug/debug"
)

func (this *TransferTo) validate() {

	debug.assert(this.fee.amount >= 0, "fee < 0")
	debug.assert(this.amount.amount > 0, "amount <= 0")

	var in, out []CommitmentType
	net_public := amount.amount.value

	for i := 0; i < len(this.outputs); i++ {
		out.append(this.outputs[i].commitment)

		if i > 0 {
			debug.assert(out[i-1] < out[i], "not in correct order")
		}
	}
	public_c := make(crypto.blind(this.bf, net_public))

	if len(outputs) > 1 {
		for out := range outputs {
			info := crypto.get_info(out.range_proof)
			debug.assert(info.max_value <= database.MAX_SUPPLY, "Exceed Max Supply")
		}
	}
}

func (this *TransferTo) calculate_fee(k FeeType) ShareType {
	return k.fee + len(this.outputs)*k.price_per_output
}

func (this *TransferFrom) validate() {
	debug.assert(this.fee.amount >= 0, "fee < 0")
	debug.assert(this.amount.amount > 0, "amount <= 0")
	debug.assert(len(this.input) > 0, "input size <= 0")
	debug.aseert(this.amount.asset_id == this.fee.asset_id, "fee must be payed by asset owner")

	var in, out CommitmentType
	net_public := this.fee.amount.value + this.amount.amount.value
	out.append(make(crypto.blind(this.bf, net_public)))

	for i := 0; i < len(this.inputs); i++ {
		in.append(this.inputs[i].commitment)
		if i > 0 {
			debug.assert(in[i-1] < in[i], "not in correct order")
		}
	}

	debug.assert(len(in) != 0, "must be at least on input")
}
