package trie

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCanUnload(t *testing.T) {
	Convey("Test CanUnload", t, func() {
		tests := []struct {
			flag                 nodeFlag
			cachegen, cachelimit uint16
			want                 bool
		}{
			{
				flag: nodeFlag{dirty: true, gen: 0},
				want: false,
			},
			{
				flag:     nodeFlag{dirty: false, gen: 0},
				cachegen: 0, cachelimit: 0,
				want: true,
			},
			{
				flag:     nodeFlag{dirty: false, gen: 65534},
				cachegen: 65535, cachelimit: 1,
				want: true,
			},
			{
				flag:     nodeFlag{dirty: false, gen: 65534},
				cachegen: 0, cachelimit: 1,
				want: true,
			},
			{
				flag:     nodeFlag{dirty: false, gen: 1},
				cachegen: 65535, cachelimit: 1,
				want: true,
			},
		}

		for _, test := range tests {
			So(test.want, ShouldEqual, test.flag.canUnload(test.cachegen, test.cachelimit))
		}
	})
}
