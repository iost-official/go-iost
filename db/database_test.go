package db

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRedisDatabase(t *testing.T) {
	Convey("Test of RedisDatabase", t, func() {
		//db, _ := DatabaseFactory("redis")
		//db, _ := DatabaseFactory("mem")
		db, _ := DatabaseFactory("ldb")

		//db, _ := NewMemDatabase()

		Convey("Put", func() {
			err := db.Put([]byte("key1"), []byte("value1"))
			So(err, ShouldBeNil)
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

func TestRedisDatabase_Hash(t *testing.T) {
	Convey("Test of type cmd", t, func() {
		db, err := DatabaseFactory("redis")
		So(err, ShouldBeNil)
		s, err := db.(*RedisDatabase).Type("iost")
		So(err, ShouldBeNil)
		So(s, ShouldEqual, "hash")

		_, err = db.(*RedisDatabase).GetAll("iost")

		So(err, ShouldBeNil)
	})
}
