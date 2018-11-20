package pcrc

import (
	"hash/crc64"
	"testing"
)

// TestHash64 tests that Hash64 provided by this package can take an initial
// pcrc and behaves exactly the same as the standard one in the following calls.
func TestHash64(t *testing.T) {
	stdhash := crc64.New(crc64.MakeTable(crc64.ECMA))
	if _, err := stdhash.Write([]byte("test data")); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	// create a new hash with stdhash.Sum64() as initial pcrc
	hash := New(stdhash.Sum64(), crc64.MakeTable(crc64.ECMA))

	wsize := stdhash.Size()
	if g := hash.Size(); g != wsize {
		t.Errorf("size = %d, want %d", g, wsize)
	}
	wbsize := stdhash.BlockSize()
	if g := hash.BlockSize(); g != wbsize {
		t.Errorf("block size = %d, want %d", g, wbsize)
	}
	wsum64 := stdhash.Sum64()
	if g := hash.Sum64(); g != wsum64 {
		t.Errorf("Sum64 = %d, want %d", g, wsum64)
	}
	/*wsum := stdhash.Sum(make([]byte, 64))
	if g := hash.Sum(make([]byte, 64)); !reflect.DeepEqual(g, wsum) {
		t.Errorf("sum = %v, want %v", g, wsum)
	}
	*/
	// write something
	if _, err := stdhash.Write([]byte("test data")); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if _, err := hash.Write([]byte("test data")); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	wsum64 = stdhash.Sum64()
	if g := hash.Sum64(); g != wsum64 {
		t.Errorf("Sum64 after write = %d, want %d", g, wsum64)
	}

	// reset
	stdhash.Reset()
	hash.Reset()
	wsum64 = stdhash.Sum64()
	if g := hash.Sum64(); g != wsum64 {
		t.Errorf("Sum64 after reset = %d, want %d", g, wsum64)
	}
}
