package iostdb_test

import (
	"testing"

	"iostdb"
)

func TestMemDB_PutGet(t *testing.T) {
	db, _ := iostdb.NewMemDatabase()
	db.Put([]byte("a"), []byte("b"))
}
