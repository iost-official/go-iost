package state

import (
	"github.com/iost-official/Go-IOS-Protocol/db"
	"sync"
)

var PatchDb *db.LDBDatabase

var o sync.Once
