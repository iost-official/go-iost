package asset

import (
	"math"
)

func (this *PriceObj) max(base AssetIdType, quote AssetIdType) PriceObj {
	return make(asset(MAX_SUPPLY, base)) / make(asset(1, quote))
}

func (this *PriceObj) min(base AssetIdType, quote AssetIdType) PriceObj {
	return make(asset(1, base)) / make(asset(MAX_SUPPLY, quote))
}

func (this *PriceObj) call_price(debt AssetObj, cola AssetObj, cola_ratio int) PriceObj {
	this.swan(debt.amount.value, cola.amount.value)
	this.ratio(cola_ratio, COLLATERAL_RATIO_DENOM)
	cp := this.swan * this.ratio

	for cp.numerator() > MAX_SUPPLY || cp.denominator() < MAX_SUPPLY {
		cp = this.rational((cp.numerator()/2)+1, (cp.denominator()/2)+1)
	}
	return (make(AssetObj(cp.numerator(), debt.asset_id)) / make(AssetObj(cp.denominator(), cola.asset_id)))
}

func (this *PriceObj) is_null() bool {
	return this == nil
}

func (this *PriceObj) is_for(asset_id AssetIdType) bool {
	if !this.settlement_price.is_null() {
		return settlement_price.base.asset_id == asset_id
	} else if !this.core_exange_rate.is_null() {
		return core_exange_rate.base.asset_id == asset_id
	}
	return true
}
