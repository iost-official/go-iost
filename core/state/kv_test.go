package state

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestKV(t *testing.T) {
	Convey("Test of values", t, func() {
		f := MakeVFloat(3.14)
		i := MakeVInt(123)
		b := MakeVByte([]byte{1, 2, 3, 4, 5})

		Convey("Test of convert and parse", func() {
			So(f.String(), ShouldEqual, "f3.140000000000000e+00")
			So(i.String(), ShouldEqual, "i123")
			So(b.String(), ShouldEqual, "bAQIDBAU=")
			m := MakeVMap(map[Key]Value{Key("f"): f, Key("i"): i, Key("b"): b,})
			So(len(m.String()), ShouldEqual, 45)

			m2, err := ParseValue("{f:f3.140000000000000e+00,i:i123,b:bAQIDBAU=,")
			So(err, ShouldBeNil)
			So(m2.Type(), ShouldEqual, Map)
		})
		Convey("Test of merge", func() {

			a, err := Merge(f, i)
			So(err, ShouldBeNil)
			So(a.String(), ShouldEqual, "i123")
			m1 := MakeVMap(map[Key]Value{Key("f"): f, Key("i"): i,})
			m2 := MakeVMap(map[Key]Value{Key("b"): b,})

			m3, err := Merge(m1, m2)
			So(err, ShouldBeNil)
			So(len(m3.String()), ShouldEqual, 45)
		})
	})
}
