package common

import "time"

const (
	SlotLength    = 3
	SecondsInHour = 3600
	SecondsInDay  = 24 * 3600
	Epoch         = 0 //1970-01-01 00:00:00
)

type Timestamp struct {
	Slot int64
}

func GetCurrentTimestamp() Timestamp {
	return GetTimestamp(time.Now().Unix())
}

func GetTimestamp(timeSec int64) Timestamp {
	return Timestamp{timeSec / SlotLength}
}

func (t *Timestamp) AddDay(intervalDay int64) {
	t.Slot = t.Slot + intervalDay*SecondsInDay/SlotLength
}

func (t *Timestamp) AddHour(intervalHour int64) {
	t.Slot = t.Slot + intervalHour*SecondsInHour/SlotLength
}

func (t *Timestamp) AddSecond(interval int64) {
	t.Slot = t.Slot + interval/SlotLength
}

func (t *Timestamp) Add(intervalSlot int64) {
	t.Slot = t.Slot + intervalSlot
}

func (t *Timestamp) ToUnixSec() int64 {
	return t.Slot*SlotLength
}

func IntervalSecond(t1 Timestamp, t2 Timestamp) int64 {
	return IntervalSecondBySlot(t1.Slot, t2.Slot)
}

func IntervalSecondBySlot(slot1 int64, slot2 int64) int64 {
	if slot1 < slot2 {
		return (slot2 - slot1) * SlotLength
	} else {
		return (slot1 - slot2) * SlotLength
	}
}

func (t1 *Timestamp) After(t2 Timestamp) bool {
	if t1.Slot <= t2.Slot {
		return true
	} else {
		return false
	}
}
