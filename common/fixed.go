package common

import (
	"encoding/binary"
	"errors"
)

// Fixed implements fixed point number for user of token balance
type Fixed struct {
	Value   int64
	Decimal int
}

// Marshal ...
func (f *Fixed) Marshal() string {
	b1 := make([]byte, 8)
	binary.LittleEndian.PutUint64(b1, uint64(f.Value))
	b2 := make([]byte, 4)
	binary.LittleEndian.PutUint32(b2, uint32(f.Decimal))
	return string(b1) + string(b2)
}

// UnmarshalFixed unmarshal from string
func UnmarshalFixed(s string) (*Fixed, error) {
	if len(s) != 8+4 {
		return &Fixed{Value: 0, Decimal: 0}, errors.New("invalid length to unmarshal fix point number")
	}
	return &Fixed{Value: int64(binary.LittleEndian.Uint64([]byte(s[:8]))), Decimal: int(int32(binary.LittleEndian.Uint32([]byte(s[8:]))))}, nil
}

// IsZero checks whether the value is zero
func (f *Fixed) IsZero() bool {
	return f.Value == 0
}

// Neg get negative number
func (f *Fixed) Neg() *Fixed {
	return &Fixed{Value: -f.Value, Decimal: f.Decimal}
}

func (f *Fixed) changeDecimal(targetDecimal int) *Fixed {
	value := f.Value
	decimal := f.Decimal
	for targetDecimal > decimal {
		decimal++
		value *= 10
	}
	for targetDecimal < decimal {
		decimal--
		value /= 10
	}
	return &Fixed{Value: value, Decimal: decimal}
}

func (f *Fixed) shrinkDecimal() *Fixed {
	value := f.Value
	decimal := f.Decimal
	for value%10 == 0 && decimal > 0 {
		value /= 10
		decimal--
	}
	return &Fixed{Value: value, Decimal: decimal}
}

// UnifyDecimal make two fix point number have same decimal.
func UnifyDecimal(a *Fixed, b *Fixed) (resultA *Fixed, resultB *Fixed) {
	if a.Decimal < b.Decimal {
		return a.changeDecimal(b.Decimal), b
	}
	return a, b.changeDecimal(a.Decimal)
}

// Equals check equal
func (f *Fixed) Equals(other *Fixed) bool {
	fpnNew, otherNew := UnifyDecimal(f, other)
	return fpnNew.Value == otherNew.Value
}

// Add ...
func (f *Fixed) Add(other *Fixed) *Fixed {
	fpnNew, otherNew := UnifyDecimal(f, other)
	return &Fixed{Value: fpnNew.Value + otherNew.Value, Decimal: fpnNew.Decimal}
}

// Sub ...
func (f *Fixed) Sub(other *Fixed) *Fixed {
	return f.Add(other.Neg())
}

// Multiply ...
func (f *Fixed) Multiply(other *Fixed) *Fixed {
	fpnNew := f.shrinkDecimal()
	otherNew := other.shrinkDecimal()
	return &Fixed{Value: fpnNew.Value * otherNew.Value, Decimal: fpnNew.Decimal + otherNew.Decimal}
}

// Times multiply a scalar
func (f *Fixed) Times(i int64) *Fixed {
	return &Fixed{Value: f.Value * i, Decimal: f.Decimal}
}

// Div divide by a scalar
func (f *Fixed) Div(i int64) *Fixed {
	return &Fixed{Value: f.Value / i, Decimal: f.Decimal}
}

// LessThan ...
func (f *Fixed) LessThan(other *Fixed) bool {
	fpnNew, otherNew := UnifyDecimal(f, other)
	return fpnNew.Value < otherNew.Value
}

// NewFixed generate Fixed from string and decimal, will truncate if decimal is smaller
func NewFixed(amount string, decimal int) (*Fixed, bool) {
	fpn := &Fixed{Value: 0, Decimal: decimal}
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

// ToString generate string of Fixed without post zero
func (f *Fixed) ToString() string {
	val := f.Value
	str := make([]byte, 0, 0)
	for val > 0 || len(str) <= f.Decimal {
		str = append(str, byte('0'+val%10))
		val /= 10
	}
	rtn := make([]byte, 0, 0)
	for i := len(str) - 1; i >= 0; i-- {
		if i+1 == f.Decimal {
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

// ToFloat ...
func (f *Fixed) ToFloat() float64 {
	return float64(f.Value) / float64(10^f.Decimal)
}
