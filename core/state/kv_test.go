package state

import (
	"testing"

	"fmt"

	. "github.com/smartystreets/goconvey/convey"
)

func TestKV(t *testing.T) {
	Convey("Test of values", t, func() {
		f := MakeVFloat(3.14)
		i := MakeVInt(123)
		b := MakeVByte([]byte{1, 2, 3, 4, 5})

		Convey("Test of convert and parse", func() {
			So(f.EncodeString(), ShouldEqual, "f3.140000000000000e+00")
			So(i.EncodeString(), ShouldEqual, "i123")
			So(b.EncodeString(), ShouldEqual, "bAQIDBAU=")
			m := MakeVMap(map[Key]Value{Key("f"): f, Key("i"): i, Key("b"): b})
			So(len(m.EncodeString()), ShouldEqual, 45)

			m2, err := ParseValue("{f:f3.140000000000000e+00,i:i123,b:bAQIDBAU=,")
			So(err, ShouldBeNil)
			So(m2.Type(), ShouldEqual, Map)
		})
		Convey("Test of merge", func() {

			a := Merge(f, i)

			So(a.EncodeString(), ShouldEqual, "i123")
			m1 := MakeVMap(map[Key]Value{Key("f"): f, Key("i"): i})
			m2 := MakeVMap(map[Key]Value{Key("b"): b})

			fmt.Println(m1.Type(), m2.Type())
			m3 := Merge(m1, m2)
			fmt.Println(m3.EncodeString())
			So(len(m3.EncodeString()), ShouldEqual, 45)
		})
	})
}
