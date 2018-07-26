package state

import (
	"sync"

	"github.com/iost-official/Go-IOS-Protocol/db"
)

var PatchDb *db.LDBDatabase

var o sync.Once
