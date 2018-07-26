package host

import (
	"testing"

	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/db"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTransfer(t *testing.T) {
	Convey("Test of transfer", t, func() {
		db, _ := db.DatabaseFactory("redis")
		mdb := state.NewDatabase(db)
		pool := state.NewPool(mdb)
		pool.PutHM("iost", "a", state.MakeVFloat(100))
		pool.PutHM("iost", "b", state.MakeVFloat(100))

		Transfer(pool, "a", "b", 20)
		aa, _ := pool.GetHM("iost", "a")
		So(aa.(*state.VFloat).ToFloat64(), ShouldEqual, 80)
		bb, _ := pool.GetHM("iost", "b")
		So(bb.(*state.VFloat).ToFloat64(), ShouldEqual, 120)

	})
}
