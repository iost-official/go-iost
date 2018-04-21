package dpos

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGlobalStaticProperty(t *testing.T) {
	Convey("Test of witness lists of static property", t, func() {
		prop := NewGlobalStaticProperty("id0", []string{"id1", "id2", "id3"})
		Convey("New", func() {
			So(prop.NumberOfWitnesses, ShouldEqual, 3)
		})

		prop.AddPendingWitness("id4")
		prop.AddPendingWitness("id5")
		Convey("Add pending witness", func() {
			So(len(prop.PendingWitnessList), ShouldEqual, 2)
		})

		Convey("Update lists", func() {
			prop.UpdateWitnessLists([]string{"id3", "id5", "id1"})
			So(prop.WitnessList[0], ShouldEqual, "id1")
			So(prop.WitnessList[1], ShouldEqual, "id3")
			So(prop.WitnessList[2], ShouldEqual, "id5")
			So(prop.PendingWitnessList[0], ShouldEqual, "id2")
			So(prop.PendingWitnessList[1], ShouldEqual, "id4")
		})
	})
}