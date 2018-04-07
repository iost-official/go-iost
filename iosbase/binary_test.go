package iosbase

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBinary(t *testing.T) {
	Convey("Test Binary", t, func() {
		Convey("slice equal of []byte", func() {
			a := []byte{1, 2, 3}
			b := []byte{1, 2, 3}
			c := []byte{2, 3, 4}
			So(Equal(a, b), ShouldBeTrue)
			So(Equal(a, c), ShouldBeFalse)
		})
		Convey("Base 58 encode and decode", func() {
			s := Base58Encode([]byte{0, 1, 255})
			So(s, ShouldEqual, "19p")
			b := Base58Decode("19p")
			So(Equal(b, []byte{0, 1, 255}), ShouldBeTrue)
		})
	})
}
