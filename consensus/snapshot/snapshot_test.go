package snapshot

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/db"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func BenchmarkSnapshot(b *testing.B) {
	os.RemoveAll("DB")
	stateDB, err := db.NewMVCCDB("DB/StateDB")

	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < 1000000; i++ {
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
	b.Run("ToSnapshot", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			err = ToSnapshot(config)
		}
	})

	b.Run("ToFile", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			err = ToFile(config)
		}
	})

}
