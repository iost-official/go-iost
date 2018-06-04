package trie

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/db"
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
)

func TestIterator(t *testing.T) {
	Convey("Test Iterator", t, func() {
		trie := newEmpty()
		vals := []struct{ k, v string }{
			{"do", "verb"},
			{"ether", "wookiedoo"},
			{"horse", "stallion"},
			{"shaman", "horse"},
			{"doge", "coin"},
			{"dog", "puppy"},
			{"somethingveryoddindeedthis is", "myothernodedata"},
		}
		all := make(map[string]string)
		for _, val := range vals {
			all[val.k] = val.v
			trie.Update([]byte(val.k), []byte(val.v))
		}
		trie.Commit(nil)

		found := make(map[string]string)
		it := NewIterator(trie.NodeIterator(nil))
		for it.Next() {
			found[string(it.Key)] = string(it.Value)
		}

		for k, v := range all {
			So(found[k], ShouldEqual, v)
		}
	})
}

type kv struct {
	k, v []byte
	t    bool
}

func LeftPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)

	return padded
}

func TestIteratorLargeData(t *testing.T) {
	Convey("Test IteratorLargedata", t, func() {
		trie := newEmpty()
		vals := make(map[string]*kv)

		for i := byte(0); i < 255; i++ {
			value := &kv{LeftPadBytes([]byte{i}, 32), []byte{i}, false}
			value2 := &kv{LeftPadBytes([]byte{10, i}, 32), []byte{i}, false}
			trie.Update(value.k, value.v)
			trie.Update(value2.k, value2.v)
			vals[string(value.k)] = value
			vals[string(value2.k)] = value2
		}

		it := NewIterator(trie.NodeIterator(nil))
		for it.Next() {
			vals[string(it.Key)].t = true
		}

		var untouched []*kv
		for _, value := range vals {
			if !value.t {
				untouched = append(untouched, value)
			}
		}

		So(len(untouched), ShouldBeLessThanOrEqualTo, 0)
	})
}

func TestNodeIteratorCoverage(t *testing.T) {
	Convey("Test NodeIteratorCoverage", t, func() {
		db, trie, _ := makeTestTrie()

		hashes := make(map[common.Hash]struct{})
		for it := trie.NodeIterator(nil); it.Next(true); {
			if it.Hash() != (common.Hash{}) {
				hashes[it.Hash()] = struct{}{}
			}
		}
		for hash := range hashes {
			_, err := db.Node(hash)
			So(err, ShouldBeNil)
		}
		for hash, obj := range db.nodes {
			if obj != nil && hash != (common.Hash{}) {
				_, ok := hashes[hash]
				So(ok, ShouldEqual, true)
			}
		}
	})
}

type kvs struct{ k, v string }

var testdata1 = []kvs{
	{"barb", "ba"},
	{"bard", "bc"},
	{"bars", "bb"},
	{"bar", "b"},
	{"fab", "z"},
	{"food", "ab"},
	{"foos", "aa"},
	{"foo", "a"},
}

var testdata2 = []kvs{
	{"aardvark", "c"},
	{"bar", "b"},
	{"barb", "bd"},
	{"bars", "be"},
	{"fab", "z"},
	{"foo", "a"},
	{"foos", "aa"},
	{"food", "ab"},
	{"jars", "d"},
}

func TestIteratorSeek(t *testing.T) {
	Convey("Test IteratorSeek", t, func() {
		trie := newEmpty()
		for _, val := range testdata1 {
			trie.Update([]byte(val.k), []byte(val.v))
		}

		// Seek to the middle.
		it := NewIterator(trie.NodeIterator([]byte("fab")))
		err := checkIteratorOrder(testdata1[4:], it)
		So(err, ShouldBeNil)

		// Seek to a non-existent key.
		it = NewIterator(trie.NodeIterator([]byte("barc")))
		err = checkIteratorOrder(testdata1[1:], it)
		So(err, ShouldBeNil)

		// Seek beyond the end.
		it = NewIterator(trie.NodeIterator([]byte("z")))
		err = checkIteratorOrder(nil, it)
		So(err, ShouldBeNil)
	})
}

func checkIteratorOrder(want []kvs, it *Iterator) error {
	for it.Next() {
		if len(want) == 0 {
			return fmt.Errorf("didn't expect any more values, got key %q", it.Key)
		}
		if !bytes.Equal(it.Key, []byte(want[0].k)) {
			return fmt.Errorf("wrong key: got %q, want %q", it.Key, want[0].k)
		}
		want = want[1:]
	}
	if len(want) > 0 {
		return fmt.Errorf("iterator ended early, want key %q", want[0])
	}
	return nil
}

func TestDifferenceIterator(t *testing.T) {
	Convey("Test DifferenceIterator", t, func() {
		triea := newEmpty()
		for _, val := range testdata1 {
			triea.Update([]byte(val.k), []byte(val.v))
		}
		triea.Commit(nil)

		trieb := newEmpty()
		for _, val := range testdata2 {
			trieb.Update([]byte(val.k), []byte(val.v))
		}
		trieb.Commit(nil)

		found := make(map[string]string)
		di, _ := NewDifferenceIterator(triea.NodeIterator(nil), trieb.NodeIterator(nil))
		it := NewIterator(di)
		for it.Next() {
			found[string(it.Key)] = string(it.Value)
		}

		all := []struct{ k, v string }{
			{"aardvark", "c"},
			{"barb", "bd"},
			{"bars", "be"},
			{"jars", "d"},
		}
		for _, item := range all {
			So(found[item.k], ShouldEqual, item.v)
		}
		So(len(found), ShouldEqual, len(all))
	})
}

func TestUnionIterator(t *testing.T) {
	Convey("Test UnionIterator", t, func() {
		triea := newEmpty()
		for _, val := range testdata1 {
			triea.Update([]byte(val.k), []byte(val.v))
		}
		triea.Commit(nil)

		trieb := newEmpty()
		for _, val := range testdata2 {
			trieb.Update([]byte(val.k), []byte(val.v))
		}
		trieb.Commit(nil)

		di, _ := NewUnionIterator([]NodeIterator{triea.NodeIterator(nil), trieb.NodeIterator(nil)})
		it := NewIterator(di)

		all := []struct{ k, v string }{
			{"aardvark", "c"},
			{"barb", "ba"},
			{"barb", "bd"},
			{"bard", "bc"},
			{"bars", "bb"},
			{"bars", "be"},
			{"bar", "b"},
			{"fab", "z"},
			{"food", "ab"},
			{"foos", "aa"},
			{"foo", "a"},
			{"jars", "d"},
		}

		for _, kv := range all {
			So(it.Next(), ShouldEqual, true)
			So(kv.k, ShouldEqual, string(it.Key))
			So(kv.v, ShouldEqual, string(it.Value))
		}
		So(it.Next(), ShouldEqual, false)
	})
}

func TestIteratorNoDups(t *testing.T) {
	Convey("Test IteratorNoDups", t, func() {
		var tr Trie
		for _, val := range testdata1 {
			tr.Update([]byte(val.k), []byte(val.v))
		}
		checkIteratorNoDups(t, tr.NodeIterator(nil), nil)
	})
}

// This test checks that nodeIterator.Next can be retried after inserting missing trie nodes.
func TestIteratorContinueAfterErrorDisk(t *testing.T)    { testIteratorContinueAfterError(t, false) }
func TestIteratorContinueAfterErrorMemonly(t *testing.T) { testIteratorContinueAfterError(t, true) }

func testIteratorContinueAfterError(t *testing.T, memonly bool) {
	Convey("Test IteratorContinueAfterError", t, func() {
		diskdb, _ := db.NewMemDatabase()
		triedb := NewDatabase(diskdb)

		tr, _ := New(common.Hash{}, triedb)
		for _, val := range testdata1 {
			tr.Update([]byte(val.k), []byte(val.v))
		}
		tr.Commit(nil)
		if !memonly {
			triedb.Commit(tr.Hash(), true)
		}
		wantNodeCount := checkIteratorNoDups(t, tr.NodeIterator(nil), nil)

		var (
			diskKeys [][]byte
			memKeys  []common.Hash
		)
		if memonly {
			memKeys = triedb.Nodes()
		} else {
			diskKeys = diskdb.Keys()
		}
		for i := 0; i < 20; i++ {
			// Create trie that will load all nodes from DB.
			tr, _ := New(tr.Hash(), triedb)

			// Remove a random node from the database. It can't be the root node
			// because that one is already loaded.
			var (
				rkey common.Hash
				rval []byte
				robj *cachedNode
			)
			for {
				if memonly {
					rkey = memKeys[rand.Intn(len(memKeys))]
				} else {
					copy(rkey[:], diskKeys[rand.Intn(len(diskKeys))])
				}
				if rkey != tr.Hash() {
					break
				}
			}
			if memonly {
				robj = triedb.nodes[rkey]
				delete(triedb.nodes, rkey)
			} else {
				rval, _ = diskdb.Get(rkey[:])
				diskdb.Delete(rkey[:])
			}
			// Iterate until the error is hit.
			seen := make(map[string]bool)
			it := tr.NodeIterator(nil)
			checkIteratorNoDups(t, it, seen)
			missing, ok := it.Error().(*MissingNodeError)
			So(ok, ShouldEqual, true)
			So(missing.NodeHash, ShouldEqual, rkey)

			// Add the node back and continue iteration.
			if memonly {
				triedb.nodes[rkey] = robj
			} else {
				diskdb.Put(rkey[:], rval)
			}
			checkIteratorNoDups(t, it, seen)
			So(it.Error(), ShouldBeNil)
			So(len(seen), ShouldEqual, wantNodeCount)
		}
	})
}

// Similar to the test above, this one checks that failure to create nodeIterator at a
// certain key prefix behaves correctly when Next is called. The expectation is that Next
// should retry seeking before returning true for the first time.
func TestIteratorContinueAfterSeekErrorDisk(t *testing.T) {
	testIteratorContinueAfterSeekError(t, false)
}
func TestIteratorContinueAfterSeekErrorMemonly(t *testing.T) {
	testIteratorContinueAfterSeekError(t, true)
}

func testIteratorContinueAfterSeekError(t *testing.T, memonly bool) {
	Convey("Test IteratorContinueAfterSeekError", t, func() {
		diskdb, _ := db.NewMemDatabase()
		triedb := NewDatabase(diskdb)

		ctr, _ := New(common.Hash{}, triedb)
		for _, val := range testdata1 {
			ctr.Update([]byte(val.k), []byte(val.v))
		}
		root, _ := ctr.Commit(nil)
		if !memonly {
			triedb.Commit(root, true)
		}
		barNodeHash := common.HexToHash("05041990364eb72fcb1127652ce40d8bab765f2bfe53225b1170d276cc101c2e")
		var (
			barNodeBlob []byte
			barNodeObj  *cachedNode
		)
		if memonly {
			barNodeObj = triedb.nodes[barNodeHash]
			delete(triedb.nodes, barNodeHash)
		} else {
			barNodeBlob, _ = diskdb.Get(barNodeHash[:])
			diskdb.Delete(barNodeHash[:])
		}
		// Create a new iterator that seeks to "bars". Seeking can't proceed because
		// the node is missing.
		tr, _ := New(root, triedb)
		it := tr.NodeIterator([]byte("bars"))
		missing, ok := it.Error().(*MissingNodeError)
		So(ok, ShouldEqual, true)
		if ok {
			So(missing.NodeHash, ShouldEqual, barNodeHash)
		}
		// Reinsert the missing node.
		if memonly {
			triedb.nodes[barNodeHash] = barNodeObj
		} else {
			diskdb.Put(barNodeHash[:], barNodeBlob)
		}
		// Check that iteration produces the right set of values.
		err := checkIteratorOrder(testdata1[2:], NewIterator(it))
		So(err, ShouldBeNil)
	})
}

func checkIteratorNoDups(t *testing.T, it NodeIterator, seen map[string]bool) int {
	if seen == nil {
		seen = make(map[string]bool)
	}
	for it.Next(true) {
		So(seen[string(it.Path())], ShouldEqual, false)
		seen[string(it.Path())] = true
	}
	return len(seen)
}
