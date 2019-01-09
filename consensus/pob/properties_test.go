package pob

import (
	"testing"

	"github.com/iost-official/go-iost/account"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGlobalStaticProperty(t *testing.T) {
	Convey("Test of witness lists of static property", t, func() {
		prop := newStaticProperty(
			&account.KeyPair{
				Pubkey: []byte{},
				Seckey: []byte{},
			},
			[]string{"id1", "id2", "id3"},
		)
		So(prop.NumberOfWitnesses, ShouldEqual, 3)
	})
}
