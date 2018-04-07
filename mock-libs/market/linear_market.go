package market

import (
	"math"

	"github.com/iost-official/PrototypeWorks/mock-libs/asset"
)

func sum_below_max_shares(a AssetObj, b AssetObj) bool {
	if a.amount > MAX_SUPPLY {
		return false
	} else if b.amount > MAX_SUPPLY {
		return false
	} else if a.amout+b.amout > MAX_SUPPLY {
		return false
	}
	return true
}

func (this *LinearMarket) get_allowed_withdraw(context PolicyObj) AssetObj {
	allowed_withdraw := 0
	if context.now > this.begin_time {
		elapsed_time := (context.now - this.begin_time).to_second()
		if elapsed_time >= this.cliff_time {
			total_vested := 0
			if elapsed_time < this.duration_time {
				total_vested = this.begin_balance.value * elapsed_time / this.duration_time
			} else {
				total_vested = this.begin_balance
			}

			withdrawn_already := this.begin_balance - context.balance.amount
			allowed_withdraw = total_vested - withdrawn_already
		}
	}
	return make(AssetObj(allowed_withdraw, context.amount.asset_id))
}

func (this *LinearMarket) is_deposit_allowed(context PlicyObj) bool {
	return context.amount.asset_id == context.balance.asset_id && sum_below_max_shares(context.amount, context.balance)
}

func (this *LinearMarket) is_withdraw_allowed(context PlicyObj) bool {
	return context.amount.asset_id == context.balance.asset_id && context.amount <= this.get_allowed_withdraw(context)
}
