package log

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestLogger(t *testing.T) {
	Convey("Test of Logger\n", t, func() {
		l, err := GetLogger("IOST", "test.log")
		So(err, ShouldBeNil)
		l.D("something %v;", "good")
		l.E("something wrong")
		l.I("something should be record")
		l.Crash("Test of Crash")
		So(true, ShouldBeTrue)
	})
}
