package common

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/iost-official/go-iost/ilog"
)

var errOverflow = errors.New("overflow error")
var errAbnormalChar = errors.New("abnormal char in amount")
var errAmountFormat = errors.New("amount format error")
var errDivideByZero = errors.New("divide by zero error")
var errDoubleDot = errors.New("double dot error")

// Fixed implements fixed point number for user of token balance
type Fixed struct {
	Value   int64
	Decimal int
	err     error
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

func multiplyOverflow(a int64, b int64) bool {
	x := a * b
	if a != 0 && x/a != b {
		return true
	}
	return false
}

// IsZero checks whether the value is zero
func (f *Fixed) IsZero() bool {
	return f.Value == 0
}

// Neg get negative number
func (f *Fixed) Neg() *Fixed {
	if multiplyOverflow(f.Value, -1) {
		f.err = errOverflow
		return nil
	}
	return &Fixed{Value: -f.Value, Decimal: f.Decimal}
}

func (f *Fixed) changeDecimal(targetDecimal int) *Fixed {
	value := f.Value
	decimal := f.Decimal
	for targetDecimal > decimal {
		decimal++
		if multiplyOverflow(value, 10) {
			f.err = errOverflow
			return nil
		}
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
func UnifyDecimal(a *Fixed, b *Fixed) (*Fixed, *Fixed, error) {
	if a.Decimal < b.Decimal {
		aChanged := a.changeDecimal(b.Decimal)
		if aChanged.err != nil {
			return nil, nil, aChanged.err
		}
		return aChanged, b, nil
	}
	bChanged := b.changeDecimal(a.Decimal)
	if bChanged.err != nil {
		return nil, nil, bChanged.err
	}
	return a, bChanged, nil
}

// Equals check equal
func (f *Fixed) Equals(other *Fixed) bool {
	fpnNew, otherNew, err := UnifyDecimal(f, other)
	f.err = err
	return fpnNew.Value == otherNew.Value
}

// Add ...
func (f *Fixed) Add(other *Fixed) *Fixed {
	fpnNew, otherNew, err := UnifyDecimal(f, other)
	if err != nil {
		f.err = err
		return nil
	}
	return &Fixed{Value: fpnNew.Value + otherNew.Value, Decimal: fpnNew.Decimal}
}

// Sub ...
func (f *Fixed) Sub(other *Fixed) *Fixed {
	ret := other.Neg()
	if other.err != nil {
		f.err = other.err
		return nil
	}
	return f.Add(ret)
}

// Multiply ...
func (f *Fixed) Multiply(other *Fixed) *Fixed {
	fpnNew := f.shrinkDecimal()
	otherNew := other.shrinkDecimal()
	if multiplyOverflow(fpnNew.Value, otherNew.Value) {
		f.err = errOverflow
		return nil
	}
	return &Fixed{Value: fpnNew.Value * otherNew.Value, Decimal: fpnNew.Decimal + otherNew.Decimal}
}

// Times multiply a scalar
func (f *Fixed) Times(i int64) *Fixed {
	if multiplyOverflow(f.Value, i) {
		f.err = errOverflow
		return nil
	}
	return &Fixed{Value: f.Value * i, Decimal: f.Decimal}
}

// Div divide by a scalar
func (f *Fixed) Div(i int64) *Fixed {
	if i == 0 {
		f.err = errDivideByZero
		return nil
	}
	return &Fixed{Value: f.Value / i, Decimal: f.Decimal}
}

// LessThan ...
func (f *Fixed) LessThan(other *Fixed) bool {
	fpnNew, otherNew, err := UnifyDecimal(f, other)
	f.err = err
	return fpnNew.Value < otherNew.Value
}

// NewFixed generate Fixed from string and decimal, will truncate if decimal is smaller
func NewFixed(amount string, decimal int) (*Fixed, error) {
	if len(amount) == 0 || amount[0] == '.' {
		return nil, errAmountFormat
	}
	if amount[0] == '-' {
		fpn, err := NewFixed(amount[1:], decimal)
		ilog.Info(fpn, err)
		if err != nil {
			return nil, err
		} else {
			return fpn.Neg(), fpn.err
		}
	}
	fpn := &Fixed{Value: 0, Decimal: 0}
	decimalStart := false
	for i := 0; i < len(amount); i++ {
		if '0' <= amount[i] && amount[i] <= '9' {
			num := int64(amount[i] - '0')
			if multiplyOverflow(fpn.Value, 10) {
				return nil, errOverflow
			}
			fpn.Value = fpn.Value*10 + num
			if fpn.Value < 0 {
				return nil, errOverflow
			}
			if decimalStart {
				fpn.Decimal++
				if fpn.Decimal >= decimal {
					break
				}
			}
		} else if amount[i] == '.' {
			if decimalStart == true {
				return nil, errDoubleDot
			}
			decimalStart = true
		} else {
			return nil, errAbnormalChar
		}
	}
	return fpn.changeDecimal(decimal), fpn.err
}

// ToString generate string of Fixed without post zero
func (f *Fixed) ToString() string {
	if f.Value < 0 {
		ret := f.Neg()
		if f.err != nil {
			return ""
		}
		str := ret.ToString()
		return "-" + str
	}
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
	if f.Decimal == 0 {
		return string(rtn)
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
	return float64(f.Value) / math.Pow10(f.Decimal)
}
