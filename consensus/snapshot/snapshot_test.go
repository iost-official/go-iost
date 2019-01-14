package snapshot

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/db"
	. "github.com/smartystreets/goconvey/convey"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func TestSnapshot(t *testing.T) {
	stateDB, err := db.NewMVCCDB("DB/StateDB")

	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < 100; i++ {
		stateDB.Put("state", randString(64), randString(32))
	}
	stateDB.Tag("abc")
	stateDB.Flush("abc")
	stateDB.Close()
	config := &common.Config{
		DB: &common.DBConfig{
			LdbPath: "DB/",
		},
	}
	defer os.RemoveAll("DB")
	Convey("Test Snapshot", t, func() {
		err = ToSnapshot(config)
		So(err, ShouldBeNil)
		err = ToFile(config)
		So(err, ShouldBeNil)

	})

}
