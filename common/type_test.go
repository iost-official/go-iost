package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNextSlot(t *testing.T) {
	assert := assert.New(t)
	SlotInterval = 1 * time.Millisecond
	for i := 0; i < 1000; i++ {
		slot := NextSlot()
		select {
		case <-time.After(time.Until(TimeOfBlock(slot, 0))):
			assert.Equal(slot, SlotOfUnixNano(time.Now().UnixNano()), "Entered the slot that was not expected.")
		}
	}
}
