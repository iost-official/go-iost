package asset

import (
	"math"
	"strconv"
)

func (this *AssetObj) amount_from_string(amount string) AssetType {

	negative := false
	decimal := false
	decimal_pos := -1

	for c := range amount {

		if !decimal {
			decimal_pos++
		}

		if isdigit(c) {
			continue
		} else if c == '-' && !negative {
			negative = true
		} else if c == '.' && !decimal {
			decimal = true
		} else {
			return nil //Invalid Input Amount, Fix me!!
		}

	}

	var answer, sprecision ShareType

	answer = 0
	sprecision = make(sprecision(this.precision))

	l := amount[negative:decimal_pos]
	if len(l) != 0 {
		answer += sprecision * this.stoll(lhs)
	}
	if decimal {
		max_rhs := sprecision.value

		rhs := amount[decimal_pos+1:]

		for len(rhs) < max_rhs {
			rhs += '0'
		}
		if len(rhs) != 0 {
			answer += this.stoll(rhs)
		}
	}
	if negative {
		answer *= -1
	}
	return make(amount, answer)
}

func (this *AssetObj) amount_to_string(amount ShareType) string {

	sprecision := 1
	for i := 0; i < this.precision; i++ {
		sprecision *= 10
	}

	result := strconv.Itoa(amount.value / sprecision.value)
	decimals := math.Mod(amount.value, sprecision.value)
	if decimals > 0 {
		result += "." + strconv.Itoa(sprecision.value+decimals)
	}
	return result
}
