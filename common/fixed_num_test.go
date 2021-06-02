package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixed_ToString(t *testing.T) {
	f := Fixed{-9223372036854775808, 4, nil}
	f.ToString()
	assert.Equal(t, f.Err, errOverflow)
	f = Fixed{100, 0, nil}
	s := f.ToString()
	assert.Equal(t, s, "100")
}

func TestNewFixed(t *testing.T) {
	f, err := NewFixed("-323.49494", 12)
	assert.Equal(t, err, nil)
	assert.Equal(t, f.Value, int64(-323494940000000))
	_, err = NewFixed("-9223372036854775808", 0)
	assert.Equal(t, err, errOverflow)
	_, err = NewFixed("-323.49494", 40)
	assert.Equal(t, err, errOverflow)
	_, err = NewFixed("323.494.94", 10)
	assert.Equal(t, err, errDoubleDot)
}

func TestFixed_Multiply(t *testing.T) {
	f1 := Fixed{-9223372036854775807, 4, nil}
	f2 := Fixed{-9223372036854775807, 4, nil}
	f1.Multiply(&f2)
	assert.Equal(t, f1.Err, errOverflow)
}

func TestFixed_Times(t *testing.T) {
	f1 := Fixed{-9223372036854775807, 4, nil}
	f1.Times(3)
	assert.Equal(t, f1.Err, errOverflow)
}

func TestFixed_Marshal(t *testing.T) {
	f, err := UnmarshalFixed((&Fixed{1230, 2, nil}).Marshal())
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, f.Decimal)
	assert.Equal(t, nil, f.Err)
	assert.Equal(t, int64(1230), f.Value)
}

func TestFixed_Compare(t *testing.T) {
	checkComp := func(a string, b string, result bool) {
		fa, err := NewFixed(a, -1)
		assert.Equal(t, err, nil)
		fb, err := NewFixed(b, -1)
		assert.Equal(t, err, nil)
		assert.Equal(t, fa.LessThan(fb), result)
	}

	// covers every branch in 'LessThan'
	// naive case
	checkComp("1.0", "999", true)
	// 1152921504606846976: 2**60
	checkComp("-1.00000001", "1152921504606846976", true)
	checkComp("-1.00000001", "-1152921504606846976", false)
	checkComp("-1152921504606846976", "-1.00000001", true)
	checkComp("0.00000000", "1152921504606846976", true)
	checkComp("0.00000000", "-1152921504606846976", false)
	checkComp("1152921504606846976", "-1.00000001", false)
	checkComp("1152921504606846976", "1.00000001", false)
	checkComp("1.00000001", "1152921504606846976", true)

}
