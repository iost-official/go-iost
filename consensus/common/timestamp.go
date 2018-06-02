// Package consensus_common contains helper functions for consensus.
package consensus_common

import "time"

const (
	SlotLength    = 3	      //一个slot设为3秒
	SecondsInHour = 3600      //一小时
	SecondsInDay  = 24 * 3600 //一天
	Epoch         = 0         //设为1970-01-01 00:00:00的Unix秒数
	//Epoch       = 1522540800 //设为2018-04-01 00:00:00的Unix秒数
)

// Timestamp 是一个slot数的wrapper
type Timestamp struct {
	Slot int64
}

// GetCurrentTimestamp 返回当前时间对应的时间戳
func GetCurrentTimestamp() Timestamp {
	t := time.Now()
	return GetTimestamp(t.Unix())
}

// GetTimestamp 根据一个Unix时间点（秒为单位）返回一个时间戳
func GetTimestamp(timeSec int64) Timestamp {
	return Timestamp{(timeSec - Epoch) / SlotLength}
}

// AddDay 在当前时间戳上增加一定天数
func (t *Timestamp) AddDay(intervalDay int) {
	t.Slot = t.Slot + int64(intervalDay)*SecondsInDay/SlotLength
}

// AddHour 在当前时间戳上增加一定小时数
func (t *Timestamp) AddHour(intervalHour int) {
	t.Slot = t.Slot + int64(intervalHour)*SecondsInHour/SlotLength
}

// AddSecond 在当前时间戳上增加一定秒数
func (t *Timestamp) AddSecond(interval int) {
	t.Slot = t.Slot + int64(interval)/SlotLength
}

// Add 在当前时间戳上增加一定slot数
func (t *Timestamp) Add(intervalSlot int) {
	t.Slot = t.Slot + int64(intervalSlot)
}

// ToUnixSec 将时间戳转化为unix秒数
func (t *Timestamp) ToUnixSec() int64 {
	return t.Slot*SlotLength + Epoch
}

// IntervalSecond 计算两个时间戳之间的秒数
func IntervalSecond(t1 Timestamp, t2 Timestamp) int64 {
	return IntervalSecondBySlot(t1.Slot, t2.Slot)
}

// IntervalSecondBySlot 计算两个slot数之间的秒数
func IntervalSecondBySlot(slot1 int64, slot2 int64) int64 {
	if slot1 < slot2 {
		return (slot2 - slot1) * SlotLength
	} else {
		return (slot1 - slot2) * SlotLength
	}
}

// After 判断两个时间戳的先后关系
func (t1 *Timestamp) After(t2 Timestamp) bool {
	if t1.Slot <= t2.Slot {
		return true
	} else {
		return false
	}
}
