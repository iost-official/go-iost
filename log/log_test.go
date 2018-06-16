package log

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLogger(t *testing.T) {
	Convey("Test of Logger\n", t, func() {
		l := Log
		l.D("something %v;", "good")
		l.E("something wrong")
		l.I("something should be record")
		//l.Crash("Test of Crash")
		So(true, ShouldBeTrue)
	})

	//Convey("Test of log existence\n", t, func() {
	//	l := Log
	//	for i := 0; i < 10; i++ {
	//		l.E("something wrong", i)
	//		time.Sleep(3 * time.Second)
	//	}
	//	So(true, ShouldBeTrue)
	//
	//})
}

//func TestOfTime(t *testing.T) {
//	ofTime()
//}
