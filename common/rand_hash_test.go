package common

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"math/rand"
)

func TestRandHash(t *testing.T) {
	// seed with 0 for deterministic behaviour
	rand.Seed(0)

	expected := []byte{0x63, 0x55, 0x62, 0x59, 0x68, 0x69, 0x5a, 0x7a}
	assert.Equal(t, expected, RandHash(8), "generate random byte array")
}
