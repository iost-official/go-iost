package state

import (
	"sync"

	"github.com/iost-official/go-iost/db"
)

var PatchDb *db.LDBDatabase

var o sync.Once
