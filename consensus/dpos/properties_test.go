package dpos

import (
	"github.com/iost-official/prototype/core"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGlobalStaticProperty(t *testing.T) {
	Convey("Test of witness lists of static property", t, func() {
		prop := NewGlobalStaticProperty(core.Member{"id0", []byte{}, []byte{}}, []string{"id1", "id2", "id3"})
		Convey("New", func() {
			So(prop.NumberOfWitnesses, ShouldEqual, 3)
		})

		prop.AddPendingWitness("id4")
		prop.AddPendingWitness("id5")
		Convey("Add pending witness", func() {
			So(len(prop.PendingWitnessList), ShouldEqual, 2)
			err := prop.AddPendingWitness("id4")
			So(err, ShouldNotBeNil)
		})

		Convey("Update lists", func() {
			prop.UpdateWitnessLists([]string{"id3", "id5", "id1"})
			So(prop.WitnessList[0], ShouldEqual, "id1")
			So(prop.WitnessList[1], ShouldEqual, "id3")
			So(prop.WitnessList[2], ShouldEqual, "id5")
			So(prop.PendingWitnessList[0], ShouldEqual, "id2")
			So(prop.PendingWitnessList[1], ShouldEqual, "id4")
		})

		Convey("Delete pending witness", func() {
			err := prop.DeletePendingWitness("id4")
			So(len(prop.PendingWitnessList), ShouldEqual, 1)
			So(err, ShouldBeNil)

			err = prop.DeletePendingWitness("id2")
			So(err, ShouldNotBeNil)
		})
	})
}
