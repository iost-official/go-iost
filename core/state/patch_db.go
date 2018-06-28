package state

import (
	"github.com/iost-official/prototype/db"
	"sync"
)

var PatchDb *db.LDBDatabase

var o sync.Once
