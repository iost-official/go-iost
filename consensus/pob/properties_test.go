package pob

import (
	"testing"
	"fmt"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGlobalStaticProperty(t *testing.T) {
	Convey("Test of witness lists of static property", t, func() {
		prop := newStaticProperty(
			account.Account{
				ID:     "id0",
				Pubkey: []byte{},
				Seckey: []byte{},
			},
			[]string{"id1", "id2", "id3"},
		)
		So(prop.NumberOfWitnesses, ShouldEqual, 3)
		prop.updateWitnessList([]string{"id3", "id4", "id5", "id6"})
		So(len(prop.WitnessList), ShouldEqual, 4)
	})
}

func TestGlobalDynamicPropertySlot(t *testing.T) {
	Convey("Test of global dynamic property (multi slot per witness)", t, func() {
		staticProperty = newStaticProperty(
			account.Account{
				ID:     "id1",
				Pubkey: []byte{},
				Seckey: []byte{},
			},
			[]string{"id0", "id1", "id2"},
		)
		startTs := common.Timestamp{Slot: 70002}
		curSec := startTs.ToUnixSec() + 1
		fmt.Println(curSec)
		sec := timeUntilNextSchedule(curSec)
		So(sec, ShouldEqual, common.SlotLength-1)
		wit := witnessOfSlot(startTs.Slot)
		So(wit, ShouldEqual, "id0")
		curSec += common.SlotLength - 1
		sec = timeUntilNextSchedule(curSec)
		So(sec, ShouldEqual, 0)
		wit = witnessOfSec(curSec)
		So(wit, ShouldEqual, "id1")
		curSec += common.SlotLength * 2
		wit = witnessOfSec(curSec)
		So(wit, ShouldEqual, "id0")
		curSec += 1
		wit = witnessOfSec(curSec)
		So(wit, ShouldEqual, "id0")
		sec = timeUntilNextSchedule(curSec)
		So(sec, ShouldEqual, common.SlotLength-1)
		curSec += common.SlotLength
		sec = timeUntilNextSchedule(curSec)
		wit = witnessOfSec(curSec)
		So(wit, ShouldEqual, "id1")
		So(sec, ShouldEqual, 0)
	})
}
