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
			3,
		)
		So(prop.NumberOfWitnesses, ShouldEqual, 3)
	})
}
