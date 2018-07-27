package state

import (
	"github.com/iost-official/Go-IOS-Protocol/db"
	"sync"
)

var PatchDb *db.LDBDatabase

var o sync.Once

//
//func init() {
//	o.Do(func() {
//		var err error
//		PatchDb, err = iostdb.NewLDBDatabase("", 0, 0)
//		if err != nil {
//			panic(err)
//		}
//	})
//}
