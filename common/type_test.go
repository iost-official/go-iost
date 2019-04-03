package common

import (
	"testing"
	"time"

	"github.com/iost-official/go-iost/ilog"
	"github.com/stretchr/testify/assert"
)

func TestNextSlotTime(t *testing.T) {
	assert := assert.New(t)
	var slotFlag int64
	SlotInterval = 1 * time.Millisecond
	for i := 0; i < 1000; i++ {
		select {
		case <-time.After(time.Until(NextSlotTime())):
			t := time.Now()
			assert.NotEqual(slotFlag, SlotOfUnixNano(t.UnixNano()), "Can't enter the same slot twice.")
			slotFlag = SlotOfUnixNano(t.UnixNano())
			ilog.Debugf("Current slot: %v", slotFlag)
		}
	}
}
