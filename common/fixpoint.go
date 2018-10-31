package common

import (
	"encoding/binary"
	"errors"
)

// FixPointNumber implements fixed point number for user of token balance
type FixPointNumber struct {
	Value   int64
	Decimal int
}

// Marshal ...
func (fpn *FixPointNumber) Marshal() string {
	b1 := make([]byte, 8)
	binary.LittleEndian.PutUint64(b1, uint64(fpn.Value))
	b2 := make([]byte, 4)
	binary.LittleEndian.PutUint32(b2, uint32(fpn.Decimal))
	return string(b1) + string(b2)
}

// UnmarshalFixPointNumber unmarshal from string
func UnmarshalFixPointNumber(s string) (*FixPointNumber, error) {
	if len(s) != 8+4 {
		return &FixPointNumber{Value: 0, Decimal: 0}, errors.New("invalid length to unmarshal fix point number")
	}
	return &FixPointNumber{Value: int64(binary.LittleEndian.Uint64([]byte(s[:8]))), Decimal: int(int32(binary.LittleEndian.Uint32([]byte(s[8:]))))}, nil
}

// Neg get negative number
func (fpn *FixPointNumber) Neg() *FixPointNumber {
	return &FixPointNumber{Value: -fpn.Value, Decimal: fpn.Decimal}
}

func (fpn *FixPointNumber) changeDecimal(targetDecimal int) *FixPointNumber {
	value := fpn.Value
	decimal := fpn.Decimal
	for targetDecimal > decimal {
		decimal++
		value *= 10
	}
	for targetDecimal < decimal {
		decimal--
		value /= 10
	}
	return &FixPointNumber{Value: value, Decimal: decimal}
}

func (fpn *FixPointNumber) shrinkDecimal() *FixPointNumber {
	value := fpn.Value
	decimal := fpn.Decimal
	for value%10 == 0 && decimal > 0 {
		value /= 10
		decimal--
	}
	return &FixPointNumber{Value: value, Decimal: decimal}
}

// UnifyDecimal make two fix point number have same decimal.
func UnifyDecimal(a *FixPointNumber, b *FixPointNumber) (resultA *FixPointNumber, resultB *FixPointNumber) {
	if a.Decimal < b.Decimal {
		return a.changeDecimal(b.Decimal), b
	}
	return a, b.changeDecimal(a.Decimal)
}

// Equals check equal
func (fpn *FixPointNumber) Equals(other *FixPointNumber) bool {
	fpnNew, otherNew := UnifyDecimal(fpn, other)
	return fpnNew.Value == otherNew.Value
}

// Add ...
func (fpn *FixPointNumber) Add(other *FixPointNumber) *FixPointNumber {
	fpnNew, otherNew := UnifyDecimal(fpn, other)
	return &FixPointNumber{Value: fpnNew.Value + otherNew.Value, Decimal: fpnNew.Decimal}
}

// Sub ...
func (fpn *FixPointNumber) Sub(other *FixPointNumber) *FixPointNumber {
	return fpn.Add(other.Neg())
}

// Multiply ...
func (fpn *FixPointNumber) Multiply(other *FixPointNumber) *FixPointNumber {
	fpnNew := fpn.shrinkDecimal()
	otherNew := other.shrinkDecimal()
	return &FixPointNumber{Value: fpnNew.Value * otherNew.Value, Decimal: fpnNew.Decimal + otherNew.Decimal}
}

// Times multiply a scalar
func (fpn *FixPointNumber) Times(i int64) *FixPointNumber {
	return &FixPointNumber{Value: fpn.Value * i, Decimal: fpn.Decimal}
}

// LessThen ...
func (fpn *FixPointNumber) LessThen(other *FixPointNumber) bool {
	fpnNew, otherNew := UnifyDecimal(fpn, other)
	return fpnNew.Value < otherNew.Value
}

// NewFixPointNumber generate FixPointNumber from string and decimal, will truncate if decimal is smaller
func NewFixPointNumber(amount string, decimal int) (*FixPointNumber, bool) {
	fpn := &FixPointNumber{Value: 0, Decimal: decimal}
	if len(amount) == 0 || amount[0] == '.' {
		return nil, false
	}
	i := 0
	for ; i < len(amount); i++ {
		if '0' <= amount[i] && amount[i] <= '9' {
			fpn.Value = fpn.Value*10 + int64(amount[i]-'0')
		} else if amount[i] == '.' {
			break
		} else {
			return nil, false
		}
	}
	for i = i + 1; i < len(amount) && decimal > 0; i++ {
		if '0' <= amount[i] && amount[i] <= '9' {
			fpn.Value = fpn.Value*10 + int64(amount[i]-'0')
			decimal = decimal - 1
		} else {
			return nil, false
		}
	}
	for decimal > 0 {
		fpn.Value = fpn.Value * 10
		decimal = decimal - 1
	}
	return fpn, true
}

// ToString generate string of FixPointNumber without post zero
func (fpn *FixPointNumber) ToString() string {
	val := fpn.Value
	str := make([]byte, 0, 0)
	for val > 0 || len(str) <= fpn.Decimal {
		str = append(str, byte('0'+val%10))
		val /= 10
	}
	rtn := make([]byte, 0, 0)
	for i := len(str) - 1; i >= 0; i-- {
		if i+1 == fpn.Decimal {
			rtn = append(rtn, '.')
		}
		rtn = append(rtn, str[i])
	}
	for rtn[len(rtn)-1] == '0' {
		rtn = rtn[0 : len(rtn)-1]
	}
	if rtn[len(rtn)-1] == '.' {
		rtn = rtn[0 : len(rtn)-1]
	}
	return string(rtn)
}
