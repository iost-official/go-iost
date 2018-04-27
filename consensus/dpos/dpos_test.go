package dpos

import (
	"testing"

	"github.com/iost-official/prototype/core"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDPoS(t *testing.T) {
	Convey("Test of DPos", t, func() {
		dpos, _ := NewDPoS(core.Member{"id0", []byte{}, []byte{}}, nil)
		dpos.Run()
	})

}
