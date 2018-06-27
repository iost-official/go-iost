// Package consensus_common contains helper functions for consensus.
package consensus_common

import "time"

const (
	SlotLength    = 3
	SecondsInHour = 3600
	SecondsInDay  = 24 * 3600
	Epoch         = 0 //1970-01-01 00:00:00
)

// Timestamp
type Timestamp struct {
	Slot int64
}

// GetCurrentTimestamp
func GetCurrentTimestamp() Timestamp {
	t := time.Now()
	return GetTimestamp(t.Unix())
}

// GetTimestamp
func GetTimestamp(timeSec int64) Timestamp {
	return Timestamp{(timeSec - Epoch) / SlotLength}
}

// AddDay
func (t *Timestamp) AddDay(intervalDay int) {
	t.Slot = t.Slot + int64(intervalDay)*SecondsInDay/SlotLength
}

// AddHour
func (t *Timestamp) AddHour(intervalHour int) {
	t.Slot = t.Slot + int64(intervalHour)*SecondsInHour/SlotLength
}

// AddSecond
func (t *Timestamp) AddSecond(interval int) {
	t.Slot = t.Slot + int64(interval)/SlotLength
}

// Add
func (t *Timestamp) Add(intervalSlot int) {
	t.Slot = t.Slot + int64(intervalSlot)
}

// ToUnixSec
func (t *Timestamp) ToUnixSec() int64 {
	return t.Slot*SlotLength + Epoch
}

// IntervalSecond
func IntervalSecond(t1 Timestamp, t2 Timestamp) int64 {
	return IntervalSecondBySlot(t1.Slot, t2.Slot)
}

// IntervalSecondBySlot
func IntervalSecondBySlot(slot1 int64, slot2 int64) int64 {
	if slot1 < slot2 {
		return (slot2 - slot1) * SlotLength
	} else {
		return (slot1 - slot2) * SlotLength
	}
}

// After
func (t1 *Timestamp) After(t2 Timestamp) bool {
	if t1.Slot <= t2.Slot {
		return true
	} else {
		return false
	}
}
