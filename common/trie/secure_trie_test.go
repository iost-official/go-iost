package trie

/*
import (
	"bytes"
	"testing"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/db"
	. "github.com/smartystreets/goconvey/convey"
	"runtime"
	"sync"
)

func newEmptySecure() *SecureTrie {
	diskdb, _ := db.NewMemDatabase()
	triedb := NewDatabase(diskdb)

	trie, _ := NewSecure(common.Hash{}, triedb, 0)
	return trie
}

// makeTestSecureTrie creates a large enough secure trie for testing.
func makeTestSecureTrie() (*Database, *SecureTrie, map[string][]byte) {
	// Create an empty trie
	diskdb, _ := db.NewMemDatabase()
	triedb := NewDatabase(diskdb)

	trie, _ := NewSecure(common.Hash{}, triedb, 0)

	// Fill it with some arbitrary data
	content := make(map[string][]byte)
	for i := byte(0); i < 255; i++ {
		// Map the same data under multiple keys
		key, val := LeftPadBytes([]byte{1, i}, 32), []byte{i}
		content[string(key)] = val
		trie.Update(key, val)

		key, val = LeftPadBytes([]byte{2, i}, 32), []byte{i}
		content[string(key)] = val
		trie.Update(key, val)

		// Add some other data to inflate the trie
		for j := byte(3); j < 13; j++ {
			key, val = LeftPadBytes([]byte{j, i}, 32), []byte{j, i}
			content[string(key)] = val
			trie.Update(key, val)
		}
	}
	trie.Commit(nil)

	// Return the generated trie
	return triedb, trie, content
}

func TestSecureDelete(t *testing.T) {
	Convey("Test SecureDelete", t, func() {
		trie := newEmptySecure()
		vals := []struct{ k, v string }{
			{"do", "verb"},
			{"ether", "wookiedoo"},
			{"horse", "stallion"},
			{"shaman", "horse"},
			{"doge", "coin"},
			{"ether", ""},
			{"dog", "puppy"},
			{"shaman", ""},
		}
		for _, val := range vals {
			if val.v != "" {
				trie.Update([]byte(val.k), []byte(val.v))
			} else {
				trie.Delete([]byte(val.k))
			}
		}
		hash := trie.Hash()
		exp := common.HexToHash("29b235a58c3c25ab83010c327d5932bcf05324b7d6b1185e650798034783ca9d")
		So(hash, ShouldEqual, exp)
	})
}

func TestSecureGetKey(t *testing.T) {
	Convey("Test SecureGet", t, func() {
		trie := newEmptySecure()
		trie.Update([]byte("foo"), []byte("bar"))

		key := []byte("foo")
		value := []byte("bar")
		seckey := Keccak256(key)

		So(true, ShouldEqual, bytes.Equal(trie.Get(key), value))
		k := trie.GetKey(seckey)
		So(true, ShouldEqual, bytes.Equal(k, key))
	})
}

func TestSecureTrieConcurrency(t *testing.T) {
	Convey("Test SecureTrieConcurrency", t, func() {
		// Create an initial trie and copy if for concurrent access
		_, trie, _ := makeTestSecureTrie()

		threads := runtime.NumCPU()
		tries := make([]*SecureTrie, threads)
		for i := 0; i < threads; i++ {
			cpy := *trie
			tries[i] = &cpy
		}
		// Start a batch of goroutines interactng with the trie
		pend := new(sync.WaitGroup)
		pend.Add(threads)
		for i := 0; i < threads; i++ {
			go func(index int) {
				defer pend.Done()

				for j := byte(0); j < 255; j++ {
					// Map the same data under multiple keys
					key, val := LeftPadBytes([]byte{byte(index), 1, j}, 32), []byte{j}
					tries[index].Update(key, val)

					key, val = LeftPadBytes([]byte{byte(index), 2, j}, 32), []byte{j}
					tries[index].Update(key, val)

					// Add some other data to inflate the trie
					for k := byte(3); k < 13; k++ {
						key, val = LeftPadBytes([]byte{byte(index), k, j}, 32), []byte{k, j}
						tries[index].Update(key, val)
					}
				}
				tries[index].Commit(nil)
			}(i)
		}
		// Wait for all threads to finish
		pend.Wait()
	})
}
*/
