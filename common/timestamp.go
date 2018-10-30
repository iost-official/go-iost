package common

import (
	"time"
)

const (
	// SlotLength interval of generate block
	SlotLength = 3
	// SecondsInHour ...
	SecondsInHour = 3600
	// SecondsInDay ...
	SecondsInDay = 24 * 3600
)

// Timestamp the interval between the blocks
type Timestamp struct {
	Slot int64
}

// GetCurrentTimestamp return value of current timestamp
func GetCurrentTimestamp() Timestamp {
	return GetTimestamp(time.Now().Unix())
}

// GetTimestamp return value of timestamp
func GetTimestamp(timeSec int64) Timestamp {
	return Timestamp{timeSec / SlotLength}
}

// AddDay add one day
func (t *Timestamp) AddDay(intervalDay int64) {
	t.Slot = t.Slot + intervalDay*SecondsInDay/SlotLength
}

// AddHour add one hour
func (t *Timestamp) AddHour(intervalHour int64) {
	t.Slot = t.Slot + intervalHour*SecondsInHour/SlotLength
}

// AddSecond add one second
func (t *Timestamp) AddSecond(interval int64) {
	t.Slot = t.Slot + interval/SlotLength
}

// Add add one slot
func (t *Timestamp) Add(intervalSlot int64) {
	t.Slot = t.Slot + intervalSlot
}

// ToUnixSec slot to second
func (t *Timestamp) ToUnixSec() int64 {
	return t.Slot * SlotLength
}

// IntervalSecond ...
func IntervalSecond(t1 Timestamp, t2 Timestamp) int64 {
	return IntervalSecondBySlot(t1.Slot, t2.Slot)
}

// IntervalSecondBySlot the interval between the slots
func IntervalSecondBySlot(slot1 int64, slot2 int64) int64 {
	if slot1 < slot2 {
		return (slot2 - slot1) * SlotLength
	}
	return (slot1 - slot2) * SlotLength
}

// After ...
func (t *Timestamp) After(t2 Timestamp) bool {
	if t.Slot <= t2.Slot {
		return true
	}
	return false
}

// ParseStringToTimestamp from time string to Timestamp
func ParseStringToTimestamp(s string) (Timestamp, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return Timestamp{0}, err
	}
	return GetTimestamp(t.Unix()), err
}
