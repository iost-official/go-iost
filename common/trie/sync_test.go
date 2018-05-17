package trie

import (
	"bytes"
	"testing"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/db"
	. "github.com/smartystreets/goconvey/convey"
)

func makeTestTrie() (*Database, *Trie, map[string][]byte) {
	// Create an empty trie
	diskdb, _ := db.NewMemDatabase()
	triedb := NewDatabase(diskdb)
	trie, _ := New(common.Hash{}, triedb)

	content := make(map[string][]byte)
	for i := byte(0); i < 255; i++ {
		key, val := LeftPadBytes([]byte{1, i}, 32), []byte{i}
		content[string(key)] = val
		trie.Update(key, val)

		key, val = LeftPadBytes([]byte{2, i}, 32), []byte{i}
		content[string(key)] = val
		trie.Update(key, val)

		for j := byte(3); j < 13; j++ {
			key, val = LeftPadBytes([]byte{j, i}, 32), []byte{j, i}
			content[string(key)] = val
			trie.Update(key, val)
		}
	}
	trie.Commit(nil)

	return triedb, trie, content
}

func checkTrieContents(t *testing.T, db *Database, root []byte, content map[string][]byte) {
	trie, err := New(common.BytesToHash(root), db)
	So(err, ShouldBeNil)
	err = checkTrieConsistency(db, common.BytesToHash(root))
	So(err, ShouldBeNil)
	for key, val := range content {
		have := trie.Get([]byte(key))
		So(bytes.Equal(have, val), ShouldEqual, true)
	}
}

func checkTrieConsistency(db *Database, root common.Hash) error {
	// Create and iterate a trie rooted in a subnode
	trie, err := New(root, db)
	if err != nil {
		return nil // Consider a non existent state consistent
	}
	it := trie.NodeIterator(nil)
	for it.Next(true) {
	}
	return it.Error()
}

// Tests that an empty trie is not scheduled for syncing.
func TestEmptyTrieSync(t *testing.T) {
	Convey("Test EmptyTrieSync", t, func() {
		diskdbA, _ := db.NewMemDatabase()
		triedbA := NewDatabase(diskdbA)

		diskdbB, _ := db.NewMemDatabase()
		triedbB := NewDatabase(diskdbB)

		emptyA, _ := New(common.Hash{}, triedbA)
		emptyB, _ := New(emptyRoot, triedbB)

		for _, trie := range []*Trie{emptyA, emptyB} {
			diskdb, _ := db.NewMemDatabase()
			req := NewTrieSync(trie.Hash(), diskdb, nil).Missing(1)
			So(len(req), ShouldEqual, 0)
		}
	})
}

// Tests that given a root hash, a trie can sync iteratively on a single thread,
// requesting retrieval tasks and returning all of them in one go.
func TestIterativeTrieSyncIndividual(t *testing.T) { testIterativeTrieSync(t, 1) }
func TestIterativeTrieSyncBatched(t *testing.T)    { testIterativeTrieSync(t, 100) }

func testIterativeTrieSync(t *testing.T, batch int) {
	Convey("Test IterativeTrieSync", t, func(){
		// Create a random trie to copy
		srcDb, srcTrie, srcData := makeTestTrie()

		// Create a destination trie and sync with the scheduler
		diskdb, _ := db.NewMemDatabase()
		triedb := NewDatabase(diskdb)
		sched := NewTrieSync(srcTrie.Hash(), diskdb, nil)

		queue := append([]common.Hash{}, sched.Missing(batch)...)
		for len(queue) > 0 {
			results := make([]SyncResult, len(queue))
			for i, hash := range queue {
				data, err := srcDb.Node(hash)
				So(err, ShouldBeNil)
				results[i] = SyncResult{hash, data}
			}
			_, _, err := sched.Process(results)
			So(err, ShouldBeNil)
			_, err = sched.Commit(diskdb)
			So(err, ShouldBeNil)
			queue = append(queue[:0], sched.Missing(batch)...)
		}
		// Cross check that the two tries are in sync
		checkTrieContents(t, triedb, srcTrie.Root(), srcData)
	})
}

func TestIterativeDelayedTrieSync(t *testing.T) {
	Convey("Test IterativeDelayedTrieSync", t, func(){
		// Create a random trie to copy
		srcDb, srcTrie, srcData := makeTestTrie()

		// Create a destination trie and sync with the scheduler
		diskdb, _ := db.NewMemDatabase()
		triedb := NewDatabase(diskdb)
		sched := NewTrieSync(srcTrie.Hash(), diskdb, nil)

		queue := append([]common.Hash{}, sched.Missing(10000)...)
		for len(queue) > 0 {
			// Sync only half of the scheduled nodes
			results := make([]SyncResult, len(queue)/2+1)
			for i, hash := range queue[:len(results)] {
				data, err := srcDb.Node(hash)
				So(err, ShouldBeNil)
				results[i] = SyncResult{hash, data}
			}
			_, _, err := sched.Process(results)
			So(err, ShouldBeNil)
			_, err = sched.Commit(diskdb)
			So(err, ShouldBeNil)
			queue = append(queue[len(results):], sched.Missing(10000)...)
		}
		// Cross check that the two tries are in sync
		checkTrieContents(t, triedb, srcTrie.Root(), srcData)
	})
}

// Tests that given a root hash, a trie can sync iteratively on a single thread,
// requesting retrieval tasks and returning all of them in one go, however in a
// random order.
func TestIterativeRandomTrieSyncIndividual(t *testing.T) { testIterativeRandomTrieSync(t, 1) }
func TestIterativeRandomTrieSyncBatched(t *testing.T)    { testIterativeRandomTrieSync(t, 100) }

func testIterativeRandomTrieSync(t *testing.T, batch int) {
	Convey("Test IterativeRandomTrieSync", t, func(){
		// Create a random trie to copy
		srcDb, srcTrie, srcData := makeTestTrie()

		// Create a destination trie and sync with the scheduler
		diskdb, _ := db.NewMemDatabase()
		triedb := NewDatabase(diskdb)
		sched := NewTrieSync(srcTrie.Hash(), diskdb, nil)

		queue := make(map[common.Hash]struct{})
		for _, hash := range sched.Missing(batch) {
			queue[hash] = struct{}{}
		}
		for len(queue) > 0 {
			// Fetch all the queued nodes in a random order
			results := make([]SyncResult, 0, len(queue))
			for hash := range queue {
				data, err := srcDb.Node(hash)
				So(err, ShouldBeNil)
				results = append(results, SyncResult{hash, data})
			}
			// Feed the retrieved results back and queue new tasks
			_, _, err := sched.Process(results)
			So(err, ShouldBeNil)
			_, err = sched.Commit(diskdb)
			So(err, ShouldBeNil)
			queue = make(map[common.Hash]struct{})
			for _, hash := range sched.Missing(batch) {
				queue[hash] = struct{}{}
			}
		}
		// Cross check that the two tries are in sync
		checkTrieContents(t, triedb, srcTrie.Root(), srcData)
	})
}

func TestIterativeRandomDelayedTrieSync(t *testing.T) {
	Convey("Test IterativeRandomDelayedTrieSync", t, func(){
		// Create a random trie to copy
		srcDb, srcTrie, srcData := makeTestTrie()

		// Create a destination trie and sync with the scheduler
		diskdb, _ := db.NewMemDatabase()
		triedb := NewDatabase(diskdb)
		sched := NewTrieSync(srcTrie.Hash(), diskdb, nil)

		queue := make(map[common.Hash]struct{})
		for _, hash := range sched.Missing(10000) {
			queue[hash] = struct{}{}
		}
		for len(queue) > 0 {
			// Sync only half of the scheduled nodes, even those in random order
			results := make([]SyncResult, 0, len(queue)/2+1)
			for hash := range queue {
				data, err := srcDb.Node(hash)
				So(err, ShouldBeNil)
				results = append(results, SyncResult{hash, data})

				if len(results) >= cap(results) {
					break
				}
			}
			// Feed the retrieved results back and queue new tasks
			_, _, err := sched.Process(results)
			So(err, ShouldBeNil)
			_, err = sched.Commit(diskdb)
			So(err, ShouldBeNil)
			for _, result := range results {
				delete(queue, result.Hash)
			}
			for _, hash := range sched.Missing(10000) {
				queue[hash] = struct{}{}
			}
		}
		// Cross check that the two tries are in sync
		checkTrieContents(t, triedb, srcTrie.Root(), srcData)
	})
}

func TestDuplicateAvoidanceTrieSync(t *testing.T) {
	Convey("Test DuplicateAvoidanceTrieSync", t, func(){
		// Create a random trie to copy
		srcDb, srcTrie, srcData := makeTestTrie()

		// Create a destination trie and sync with the scheduler
		diskdb, _ := db.NewMemDatabase()
		triedb := NewDatabase(diskdb)
		sched := NewTrieSync(srcTrie.Hash(), diskdb, nil)

		queue := append([]common.Hash{}, sched.Missing(0)...)
		requested := make(map[common.Hash]struct{})

		for len(queue) > 0 {
			results := make([]SyncResult, len(queue))
			for i, hash := range queue {
				data, err := srcDb.Node(hash)
				So(err, ShouldBeNil)
				_, ok := requested[hash]
				So(ok, ShouldEqual, false)
				requested[hash] = struct{}{}

				results[i] = SyncResult{hash, data}
			}
			_, _, err := sched.Process(results)
			So(err, ShouldBeNil)
			_, err = sched.Commit(diskdb)
			So(err, ShouldBeNil)
			queue = append(queue[:0], sched.Missing(0)...)
		}
		// Cross check that the two tries are in sync
		checkTrieContents(t, triedb, srcTrie.Root(), srcData)
	})
}

func TestIncompleteTrieSync(t *testing.T) {
	Convey("Test IncompleteTrieSync", t, func(){
		// Create a random trie to copy
		srcDb, srcTrie, _ := makeTestTrie()

		diskdb, _ := db.NewMemDatabase()
		triedb := NewDatabase(diskdb)
		sched := NewTrieSync(srcTrie.Hash(), diskdb, nil)

		added := []common.Hash{}
		queue := append([]common.Hash{}, sched.Missing(1)...)
		for len(queue) > 0 {
			// Fetch a batch of trie nodes
			results := make([]SyncResult, len(queue))
			for i, hash := range queue {
				data, err := srcDb.Node(hash)
				So(err, ShouldBeNil)
				results[i] = SyncResult{hash, data}
			}
			// Process each of the trie nodes
			_, _, err := sched.Process(results)
			So(err, ShouldBeNil)
			_, err = sched.Commit(diskdb)
			So(err, ShouldBeNil)
			for _, result := range results {
				added = append(added, result.Hash)
			}
			// Check that all known sub-tries in the synced trie are complete
			for _, root := range added {
				err := checkTrieConsistency(triedb, root)
				So(err, ShouldBeNil)
			}
			// Fetch the next batch to retrieve
			queue = append(queue[:0], sched.Missing(1)...)
		}
		// Sanity check that removing any node from the database is detected
		for _, node := range added[1:] {
			key := node.Bytes()
			value, _ := diskdb.Get(key)

			diskdb.Delete(key)
			err := checkTrieConsistency(triedb, added[0])
			So(err, ShouldNotBeNil)
			diskdb.Put(key, value)
		}
	})
}
