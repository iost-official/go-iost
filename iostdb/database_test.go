package iostdb

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRedisDatabase(t *testing.T) {
	Convey("Test of RedisDatabase", t, func() {
		db, _ := NewRedisDatabase()
		Convey("Put", func() {
			err := db.Put("key", "field1", "value1", "field2", "value2")
			So(err, ShouldBeNil)
		})
		Convey("Get", func() {
			rtn, _ := db.Get("key", "field1", "field2")
			fmt.Println(rtn)
		})
		Convey("Has", func() {
			rtn, _ := db.Has("key")
			So(rtn, ShouldBeTrue)
		})
		Convey("Del", func() {
			err := db.Delete("key")
			So(err, ShouldBeNil)
		})
		Convey("Get_2", func() {
			rtn, _ := db.Get("key", "field1", "field2")
			fmt.Println(rtn)
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
