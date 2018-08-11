package db

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDatabase(t *testing.T) {
	Convey("Test of Database", t, func() {
		db, _ := NewLDB("LDB", 0, 0)

		Convey("Put", func() {
			btch := db.Batch()
			err := btch.Put([]byte("key1"), []byte("value1"))
			So(err, ShouldBeNil)
			btch.Commit()
		})

		Convey("Get", func() {
			//			db.Put([]byte("key1"), []byte("value1"))
			rtn, _ := db.Get([]byte("key1"))
			So(string(rtn), ShouldEqual, "value1")
		})
		Convey("Has", func() {
			db.Put([]byte("key1"), []byte("value1"))
			rtn, _ := db.Has([]byte("key1"))
			So(rtn, ShouldBeTrue)
		})

		Convey("Del&Get", func() {
			db.Put([]byte("key1"), []byte("value1"))
			err := db.Delete([]byte("key1"))
			So(err, ShouldBeNil)
			rtn, _ := db.Get([]byte("key1"))
			So(rtn, ShouldBeNil)
		})
	})
}
