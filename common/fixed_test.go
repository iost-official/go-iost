package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixed_ToString(t *testing.T) {
	f := Fixed{-9223372036854775808, 4}
	_, err := f.ToString()
	assert.Equal(t, err, errOverflow)
}

func TestNewFixed(t *testing.T) {
	f, err := NewFixed("-323.49494", 12)
	assert.Equal(t, err, nil)
	assert.Equal(t, f.Value, int64(-323494940000000))
	_, err = NewFixed("-9223372036854775808", 0)
	assert.Equal(t, err, errOverflow)
	_, err = NewFixed("-323.49494", 40)
	assert.Equal(t, err, errOverflow)
}

func TestFixed_Multiply(t *testing.T) {
	f1 := Fixed{-9223372036854775807, 4}
	f2 := Fixed{-9223372036854775807, 4}
	_, err := f1.Multiply(&f2)
	assert.Equal(t, err, errOverflow)
}

func TestFixed_Times(t *testing.T) {
	f1 := Fixed{-9223372036854775807, 4}
	_, err := f1.Times(3)
	assert.Equal(t, err, errOverflow)
}
