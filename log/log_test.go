package log

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLogger(t *testing.T) {
	Convey("Test of Logger\n", t, func() {
		l := NewLogger("IOST", "")
		l.D("something %v;", "good")
		l.E("something wrong")
		l.I("something should be record")
		l.Crash("Test of Crash")
		So(true, ShouldBeTrue)
	})
}