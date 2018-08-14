package pob

import (
	"testing"

	"github.com/iost-official/Go-IOS-Protocol/account"
	. "github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/iost-official/Go-IOS-Protocol/common"
)

func TestGlobalStaticProperty(t *testing.T) {
	Convey("Test of witness lists of static property", t, func() {
		prop := newGlobalStaticProperty(
			Account{
				ID:     "id0",
				Pubkey: []byte{},
				Seckey: []byte{},
			},
			[]string{"id1", "id2", "id3"},
		)
		Convey("New", func() {
			So(prop.NumberOfWitnesses, ShouldEqual, 3)
		})

		prop.updateWitnessList([]string{"id3", "id4", "id5", "id6"})
		Convey("Update pending witness", func() {
			So(len(prop.WitnessList), ShouldEqual, 4)
		})

		prop.addSlotWitness(1, "id3")
		prop.addSlotWitness(1, "id4")
		prop.addSlotWitness(2, "id5")
		prop.addSlotWitness(3, "id6")

		Convey("Slot map has", func() {
			So(prop.hasSlotWitness(2, "id5"), ShouldBeTrue)
			So(prop.hasSlotWitness(2, "id6"), ShouldBeFalse)
			So(prop.hasSlotWitness(4, "id2"), ShouldBeFalse)
		})

		Convey("Slot map add and delete", func() {
			prop.addSlotWitness(4, "id3")
			So(prop.hasSlotWitness(4, "id3"), ShouldBeTrue)
			prop.delSlotWitness(0, 2)
			So(prop.hasSlotWitness(1, "id3"), ShouldBeFalse)
			So(prop.hasSlotWitness(2, "id5"), ShouldBeFalse)
			So(prop.hasSlotWitness(3, "id6"), ShouldBeTrue)
		})
	})
}

func TestGlobalDynamicPropertyOneSlot(t *testing.T) {
	Convey("Test of global dynamic property (one slot per witness)", t, func() {
		staticProperty = newGlobalStaticProperty(
			account.Account{
				ID:     "id1",
				Pubkey: []byte{},
				Seckey: []byte{},
			},
			[]string{"id0", "id1", "id2"},
		)

		dynamicProperty = newGlobalDynamicProperty()
		dynamicProperty.LastBlockNumber = 0
		dynamicProperty.TotalSlots = 0
		dynamicProperty.LastBlockTime = common.Timestamp{Slot: 0}
		startTs := common.Timestamp{Slot: 70002}
		bh := block.BlockHead{
			Number:  1,
			Time:    startTs.Slot,
			Witness: "id0",
		}
		dynamicProperty.update(&block.Block{Head:bh})

		Convey("update first block", func() {
			So(dynamicProperty.LastBlockNumber, ShouldEqual, 1)
		})

		curSec := startTs.ToUnixSec() + 1
		sec := timeUntilNextSchedule(curSec)
		Convey("in other's slot", func() {
			curTs := common.GetTimestamp(curSec)
			wit := witnessOfTime(curTs)
			So(wit, ShouldEqual, "id0")
			So(sec, ShouldBeLessThanOrEqualTo, common.SlotLength)
		})

		curSec += common.SlotLength - 1

		timestamp := common.GetTimestamp(curSec)
		Convey("in self's slot", func() {
			wit := witnessOfTime(timestamp)
			So(wit, ShouldEqual, "id1")
			wit = witnessOfSec(curSec)
			So(wit, ShouldEqual, "id1")
		})

		bh.Number = 2
		bh.Time = timestamp.Slot
		bh.Witness = "id1"
		//dynamicProperty.update(&bh)
		Convey("update second block", func() {
			So(dynamicProperty.LastBlockNumber, ShouldEqual, 2)
		})

		curSec += 1
		sec = timeUntilNextSchedule(curSec)
		Convey("in self's slot, but finished", func() {
			So(sec, ShouldBeGreaterThanOrEqualTo, common.SlotLength*2)
			So(sec, ShouldBeLessThanOrEqualTo, common.SlotLength*3)
		})

		curSec += common.SlotLength*3 - 1
		Convey("in self's slot and lost two previous blocks", func() {
			curTs := common.GetTimestamp(curSec)
			wit := witnessOfTime(curTs)
			So(wit, ShouldEqual, "id1")
			wit = witnessOfSec(curSec)
			So(wit, ShouldEqual, "id1")
		})

		timestamp = common.GetTimestamp(curSec)
		bh.Number = 3
		bh.Time = timestamp.Slot
		bh.Witness = "id1"
		//dynamicProperty.update(&bh)
		Convey("update third block", func() {
			So(dynamicProperty.LastBlockNumber, ShouldEqual, 3)
		})
	})
}

func TestGlobalDynamicPropertyMultiSlot(t *testing.T) {
	Convey("Test of global dynamic property (multi slot per witness)", t, func() {
		slotPerWitness = 3
		staticProperty = newGlobalStaticProperty(
			account.Account{
				ID:     "id1",
				Pubkey: []byte{},
				Seckey: []byte{},
			},
			[]string{"id0", "id1", "id2"},
		)

		dynamicProperty = newGlobalDynamicProperty()
		dynamicProperty.LastBlockNumber = 0
		dynamicProperty.TotalSlots = 0
		dynamicProperty.LastBlockTime = common.Timestamp{Slot: 0}
		startTs := common.Timestamp{Slot: 70002}
		bh := block.BlockHead{
			Number:  1,
			Time:    startTs.Slot,
			Witness: "id0",
		}
		//dynamicProperty.update(&bh)

		Convey("update first block", func() {
			So(dynamicProperty.LastBlockNumber, ShouldEqual, 1)
		})

		curSec := startTs.ToUnixSec() + 1
		sec := timeUntilNextSchedule(curSec)
		Convey("in other's slot", func() {
			curTs := common.GetTimestamp(curSec)
			wit := witnessOfTime(curTs)
			So(wit, ShouldEqual, "id0")
			So(sec, ShouldEqual, 3*common.SlotLength-1)
		})

		curSec += common.SlotLength - 1
		timestamp := common.GetTimestamp(curSec)
		sec = timeUntilNextSchedule(curSec)
		Convey("in other's second slot", func() {
			wit := witnessOfTime(timestamp)
			So(wit, ShouldEqual, "id0")
			So(sec, ShouldEqual, 2*common.SlotLength)
		})

		bh.Number = 2
		bh.Time = timestamp.Slot
		bh.Witness = "id0"
		//dynamicProperty.update(&bh)
		Convey("update second block", func() {
			So(dynamicProperty.LastBlockNumber, ShouldEqual, 2)
		})

		curSec += common.SlotLength * 2
		Convey("in self's slot", func() {
			wit := witnessOfSec(curSec)
			So(wit, ShouldEqual, "id1")
		})
		bh.Number = 3
		bh.Time = common.GetTimestamp(curSec).Slot
		bh.Witness = "id1"
		//dynamicProperty.update(&bh)

		curSec += 1
		sec = timeUntilNextSchedule(curSec)
		Convey("in self's first slot, but finished", func() {
			wit := witnessOfSec(curSec)
			So(wit, ShouldEqual, "id1")
			So(sec, ShouldEqual, common.SlotLength-1)
		})

		curSec += common.SlotLength
		sec = timeUntilNextSchedule(curSec)
		Convey("in self's second slot, not on time", func() {
			wit := witnessOfSec(curSec)
			So(wit, ShouldEqual, "id1")
			So(sec, ShouldEqual, 0)
		})
		bh.Number = 4
		bh.Time = common.GetTimestamp(curSec).Slot
		bh.Witness = "id1"
		//dynamicProperty.update(&bh)

		curSec += common.SlotLength*2 - 1
		sec = timeUntilNextSchedule(curSec)
		Convey("past self's last slot", func() {
			wit := witnessOfSec(curSec)
			So(wit, ShouldEqual, "id2")
			So(sec, ShouldEqual, 6*common.SlotLength)
		})
	})
}
