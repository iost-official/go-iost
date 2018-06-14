package host

import (
	"fmt"
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
		fmt.Println(pool.GetHM("iost", "a"))
		fmt.Println(pool.GetHM("iost", "b"))

	})
}
