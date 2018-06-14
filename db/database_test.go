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
		/*
			Convey("PutHM", func() {
				err := db.PutHM([]byte("key"), []byte("field1"), []byte("value1"), []byte("field2"), []byte("value2"))
				So(err, ShouldBeNil)
			})

			Convey("GetHM", func() {
				db.PutHM([]byte("key"), []byte("field1"), []byte("value1"), []byte("field2"), []byte("value2"))
				rtn, err := db.GetHM([]byte("key"), []byte("field1"), []byte("field2"))
				So(err, ShouldBeNil)
				if err == nil {
					So(string(rtn[0]), ShouldEqual, "value1")
					So(string(rtn[1]), ShouldEqual, "value2")
				}
			})
		*/
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
	/*
		Convey("Test of UTXORedis", t, func() {
			db, _ := NewUTXORedis("value1")
			Convey("Put", func() {
				err := db.Put("key", "1")
				So(err, ShouldBeNil)
			})
			Convey("Get", func() {
				rtn, _ := db.Get("key")
				fmt.Println(rtn.([]interface{}))
			})
			Convey("Has", func() {
				rtn, _ := db.Has("key")
				So(rtn, ShouldBeTrue)
			})

		})*/
}
