package dpos

import (
	"testing"

	"github.com/iost-official/prototype/core"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDPoS(t *testing.T) {
	Convey("Test of DPos", t, func() {
		dpos := NewDPoS(core.Member{"id0", []byte{}, []byte{}}, []string{"id1", "id2", "id3"}, core.BlockChain{})

	})

}
