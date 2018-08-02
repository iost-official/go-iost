package pob

import (
	"testing"

	"github.com/iost-official/Go-IOS-Protocol/account"
	. "github.com/iost-official/Go-IOS-Protocol/account"
	. "github.com/iost-official/Go-IOS-Protocol/consensus/common"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	. "github.com/smartystreets/goconvey/convey"
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
			prop.delSlotWitness(0,2)
			So(prop.hasSlotWitness(1, "id3"), ShouldBeFalse)
			So(prop.hasSlotWitness(2, "id5"), ShouldBeFalse)
			So(prop.hasSlotWitness(3, "id6"), ShouldBeTrue)
		})
	})
}

func TestGlobalDynamicProperty(t *testing.T) {
	Convey("Test of global dynamic property", t, func() {
		staticProp = newGlobalStaticProperty(
			account.Account{
				ID:     "id1",
				Pubkey: []byte{},
				Seckey: []byte{},
			},
			[]string{"id0", "id1", "id2"},
		)

		dynamicProp = newGlobalDynamicProperty()
		dynamicProp.LastBlockNumber = 0
		dynamicProp.TotalSlots = 0
		dynamicProp.LastBlockTime = Timestamp{Slot: 0}
		startTs := Timestamp{Slot: 70002}
		bh := block.BlockHead{
			Number:  1,
			Time:    startTs.Slot,
			Witness: "id0",
		}
		dynamicProp.update(&bh)

		Convey("update first block", func() {
			So(dynamicProp.LastBlockNumber, ShouldEqual, 1)
		})

		curSec := startTs.ToUnixSec() + 1
		sec := timeUntilNextSchedule(curSec)
		Convey("in other's slot", func() {
			curTs := GetTimestamp(curSec)
			wit := witnessOfTime(curTs)
			So(wit, ShouldEqual, "id0")
			So(sec, ShouldBeLessThanOrEqualTo, SlotLength)
		})

		curSec += SlotLength - 1

		timestamp := GetTimestamp(curSec)
		Convey("in self's slot", func() {
			wit := witnessOfTime(timestamp)
			So(wit, ShouldEqual, "id1")
			wit = witnessOfSec(curSec)
			So(wit, ShouldEqual, "id1")
		})

		bh.Number = 2
		bh.Time = timestamp.Slot
		bh.Witness = "id1"
		dynamicProp.update(&bh)
		Convey("update second block", func() {
			So(dynamicProp.LastBlockNumber, ShouldEqual, 2)
		})

		curSec += 1
		sec = timeUntilNextSchedule(curSec)
		Convey("in self's slot, but finished", func() {
			So(sec, ShouldBeGreaterThanOrEqualTo, SlotLength*2)
			So(sec, ShouldBeLessThanOrEqualTo, SlotLength*3)
		})

		curSec += SlotLength*3 - 1
		Convey("in self's slot and lost two previous blocks", func() {
			curTs := GetTimestamp(curSec)
			wit := witnessOfTime(curTs)
			So(wit, ShouldEqual, "id1")
			wit = witnessOfSec(curSec)
			So(wit, ShouldEqual, "id1")
		})

		timestamp = GetTimestamp(curSec)
		bh.Number = 3
		bh.Time = timestamp.Slot
		bh.Witness = "id1"
		dynamicProp.update(&bh)
		Convey("update third block", func() {
			So(dynamicProp.LastBlockNumber, ShouldEqual, 3)
		})
	})
}
