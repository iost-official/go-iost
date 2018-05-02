package core

import "time"

const (
	SlotLength    = 3 //一个slot设为3秒
	SecondsInHour = 3600
	SecondsInDay  = 24 * 3600
	Epoch         = 1522540800 //设为2018-04-01 00:00:00的Unix秒数
)

// 返回当前时间对应的时间戳
func GetCurrentTimestamp() Timestamp {
	t := time.Now()
	return GetTimestamp(t.Unix())
}

// 根据一个Unix时间点（秒为单位）返回一个时间戳
func GetTimestamp(timeSec int64) Timestamp {
	return Timestamp{(timeSec - Epoch) / SlotLength}
}

func (t *Timestamp) AddDay(intervalDay int) {
	t.Slot = t.Slot + int64(intervalDay)*SecondsInDay/SlotLength
}

func (t *Timestamp) AddHour(intervalHour int) {
	t.Slot = t.Slot + int64(intervalHour)*SecondsInHour/SlotLength
}

func (t *Timestamp) AddSecond(interval int) {
	t.Slot = t.Slot + int64(interval)/SlotLength
}

func (t *Timestamp) Add(intervalSlot int) {
	t.Slot = t.Slot + int64(intervalSlot)
}

func (t *Timestamp) ToUnixSec() int64 {
	return t.Slot*SlotLength + Epoch
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
