package iostdb_test

import (
	"testing"

	"iostdb"
)

var test_values = []string{"", "a", "123", "\x00"}

func TestMemDB_PutGet(t *testing.T) {
	db, _ := iostdb.NewMemDatabase()
	testPutGet(db, t)
}

func testPutGet(db iostdb.Database, t *testing.T) {
	t.Parallel()
	for _, v := range test_values {
		err := db.Put([]byte(v), []byte(v))
		if err != nil {
			t.Fatalf("put failed: %v", err)
		}
	}
}
