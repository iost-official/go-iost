package trie

/*
import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/common/rlp"
	"github.com/iost-official/prototype/db"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"testing/quick"
)

func init() {
	spew.Config.Indent = "    "
	spew.Config.DisableMethods = false
}

// Used for testing
func newEmpty() *Trie {
	diskdb, _ := db.NewMemDatabase()
	trie, _ := New(common.Hash{}, NewDatabase(diskdb))
	return trie
}

func TestEmptyTrie(t *testing.T) {
	Convey("Test EmptyTrie", t, func() {
		var trie Trie
		res := trie.Hash()
		exp := emptyRoot
		So(res, ShouldEqual, common.Hash(exp))
	})
}

func TestNull(t *testing.T) {
	Convey("Test Null", t, func() {
		var trie Trie
		key := make([]byte, 32)
		value := []byte("test")
		trie.Update(key, value)
		So(true, ShouldEqual, bytes.Equal(trie.Get(key), value))
	})
}

func TestMissingRoot(t *testing.T) {
	Convey("Test MissingRoot", t, func() {
		diskdb, _ := db.NewMemDatabase()
		trie, err := New(common.HexToHash("0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33"), NewDatabase(diskdb))
		if trie != nil {
			t.Error("New returned non-nil trie for invalid root")
		}
		_, ok := err.(*MissingNodeError)
		So(trie, ShouldBeNil)
		So(true, ShouldEqual, ok)
	})
}

func TestMissingNodeDisk(t *testing.T)    { testMissingNode(t, false) }
func TestMissingNodeMemonly(t *testing.T) { testMissingNode(t, true) }

func testMissingNode(t *testing.T, memonly bool) {
	Convey("Test MissingNode", t, func() {
		diskdb, _ := db.NewMemDatabase()
		triedb := NewDatabase(diskdb)

		trie, _ := New(common.Hash{}, triedb)
		updateString(trie, "120000", "qwerqwerqwerqwerqwerqwerqwerqwer")
		updateString(trie, "123456", "asdfasdfasdfasdfasdfasdfasdfasdf")
		root, _ := trie.Commit(nil)
		if !memonly {
			triedb.Commit(root, true)
		}

		trie, _ = New(root, triedb)
		_, err := trie.TryGet([]byte("120000"))
		So(err, ShouldEqual, nil)
		trie, _ = New(root, triedb)
		_, err = trie.TryGet([]byte("120099"))
		So(err, ShouldEqual, nil)
		trie, _ = New(root, triedb)
		_, err = trie.TryGet([]byte("123456"))
		So(err, ShouldEqual, nil)
		trie, _ = New(root, triedb)
		err = trie.TryUpdate([]byte("120099"), []byte("zxcvzxcvzxcvzxcvzxcvzxcvzxcvzxcv"))
		So(err, ShouldEqual, nil)
		trie, _ = New(root, triedb)
		err = trie.TryDelete([]byte("123456"))
		So(err, ShouldEqual, nil)

		hash := common.HexToHash("0xe1d943cc8f061a0c0b98162830b970395ac9315654824bf21b73b891365262f9")
		if memonly {
			delete(triedb.nodes, hash)
		} else {
			diskdb.Delete(hash[:])
		}

		trie, _ = New(root, triedb)
		_, err = trie.TryGet([]byte("120000"))
		_, ok := err.(*MissingNodeError)
		So(true, ShouldEqual, ok)
		trie, _ = New(root, triedb)
		_, err = trie.TryGet([]byte("120099"))
		_, ok = err.(*MissingNodeError)
		So(true, ShouldEqual, ok)
		trie, _ = New(root, triedb)
		_, err = trie.TryGet([]byte("123456"))
		So(err, ShouldEqual, nil)
		trie, _ = New(root, triedb)
		err = trie.TryUpdate([]byte("120099"), []byte("zxcv"))
		_, ok = err.(*MissingNodeError)
		So(true, ShouldEqual, ok)
		trie, _ = New(root, triedb)
		err = trie.TryDelete([]byte("123456"))
		_, ok = err.(*MissingNodeError)
		So(true, ShouldEqual, ok)
	})
}

func TestInsert(t *testing.T) {
	Convey("Test Insert", t, func() {
		trie := newEmpty()

		updateString(trie, "doe", "reindeer")
		updateString(trie, "dog", "puppy")
		updateString(trie, "dogglesworth", "cat")

		exp := common.HexToHash("8aad789dff2f538bca5d8ea56e8abe10f4c7ba3a5dea95fea4cd6e7c3a1168d3")
		root := trie.Hash()
		So(root, ShouldEqual, exp)

		trie = newEmpty()
		updateString(trie, "A", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

		exp = common.HexToHash("d23786fb4a010da3ce639d66d5e904a11dbc02746d1ce25029e53290cabf28ab")
		root, err := trie.Commit(nil)
		So(err, ShouldEqual, nil)
		So(root, ShouldEqual, exp)
	})
}

func TestGet(t *testing.T) {
	Convey("Test Get", t, func() {
		trie := newEmpty()
		updateString(trie, "doe", "reindeer")
		updateString(trie, "dog", "puppy")
		updateString(trie, "dogglesworth", "cat")

		for i := 0; i < 2; i++ {
			res := getString(trie, "dog")
			So(true, ShouldEqual, bytes.Equal(res, []byte("puppy")))

			unknown := getString(trie, "unknown")
			So(unknown, ShouldEqual, nil)

			if i == 1 {
				return
			}
			trie.Commit(nil)
		}
	})
}

func TestDelete(t *testing.T) {
	Convey("Test Delete", t, func() {
		trie := newEmpty()
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
				updateString(trie, val.k, val.v)
			} else {
				deleteString(trie, val.k)
			}
		}

		hash := trie.Hash()
		exp := common.HexToHash("5991bb8c6514148a29db676a14ac506cd2cd5775ace63c30a4fe457715e9ac84")
		So(hash, ShouldEqual, exp)
	})
}

func TestEmptyValues(t *testing.T) {
	Convey("Test EmptyValues", t, func() {
		trie := newEmpty()

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
			updateString(trie, val.k, val.v)
		}

		hash := trie.Hash()
		exp := common.HexToHash("5991bb8c6514148a29db676a14ac506cd2cd5775ace63c30a4fe457715e9ac84")
		So(hash, ShouldEqual, exp)
	})
}

func TestReplication(t *testing.T) {
	Convey("Test Replication", t, func() {
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
		for _, val := range vals {
			updateString(trie, val.k, val.v)
		}
		exp, err := trie.Commit(nil)
		So(err, ShouldEqual, nil)

		// create a new trie on top of the database and check that lookups work.
		trie2, err := New(exp, trie.db)
		So(err, ShouldEqual, nil)
		for _, kv := range vals {
			So(kv.v, ShouldEqual, string(getString(trie2, kv.k)))
		}
		hash, err := trie2.Commit(nil)
		So(err, ShouldEqual, nil)
		So(hash, ShouldEqual, exp)

		// perform some insertions on the new trie.
		vals2 := []struct{ k, v string }{
			{"do", "verb"},
			{"ether", "wookiedoo"},
			{"horse", "stallion"},
		}
		for _, val := range vals2 {
			updateString(trie2, val.k, val.v)
		}
		hash = trie2.Hash()
		So(hash, ShouldEqual, exp)
	})
}

func TestLargeValue(t *testing.T) {
	Convey("Test LargeValue", t, func() {
		trie := newEmpty()
		trie.Update([]byte("key1"), []byte{99, 99, 99, 99})
		trie.Update([]byte("key2"), bytes.Repeat([]byte{1}, 32))
		trie.Hash()
	})
}

type countingDB struct {
	db.Database
	gets map[string]int
}

func (db *countingDB) Get(key []byte) ([]byte, error) {
	db.gets[string(key)]++
	return db.Database.Get(key)
}

func TestCacheUnload(t *testing.T) {
	Convey("Test CacheUnload", t, func() {
		// Create test trie with two branches.
		trie := newEmpty()
		key1 := "---------------------------------"
		key2 := "---some other branch"
		updateString(trie, key1, "this is the branch of key1.")
		updateString(trie, key2, "this is the branch of key2.")

		root, _ := trie.Commit(nil)
		trie.db.Commit(root, true)

		db := &countingDB{Database: trie.db.diskdb, gets: make(map[string]int)}
		trie, _ = New(root, NewDatabase(db))
		trie.SetCacheLimit(5)
		for i := 0; i < 12; i++ {
			getString(trie, key1)
			trie.Commit(nil)
		}
		// Check that it got loaded two times.
		for _, count := range db.gets {
			So(count, ShouldEqual, 2)
		}
	})
}

type randTest []randTestStep

type randTestStep struct {
	op    int
	key   []byte
	value []byte
	err   error
}

const (
	opUpdate = iota
	opDelete
	opGet
	opCommit
	opHash
	opReset
	opItercheckhash
	opCheckCacheInvariant
	opMax // boundary value, not an actual op
)

func (randTest) Generate(r *rand.Rand, size int) reflect.Value {
	var allKeys [][]byte
	genKey := func() []byte {
		if len(allKeys) < 2 || r.Intn(100) < 10 {
			// new key
			key := make([]byte, r.Intn(50))
			r.Read(key)
			allKeys = append(allKeys, key)
			return key
		}
		// use existing key
		return allKeys[r.Intn(len(allKeys))]
	}

	var steps randTest
	for i := 0; i < size; i++ {
		step := randTestStep{op: r.Intn(opMax)}
		switch step.op {
		case opUpdate:
			step.key = genKey()
			step.value = make([]byte, 8)
			binary.BigEndian.PutUint64(step.value, uint64(i))
		case opGet, opDelete:
			step.key = genKey()
		}
		steps = append(steps, step)
	}
	return reflect.ValueOf(steps)
}

func runRandTest(rt randTest) bool {
	diskdb, _ := db.NewMemDatabase()
	triedb := NewDatabase(diskdb)

	tr, _ := New(common.Hash{}, triedb)
	values := make(map[string]string) // tracks content of the trie

	for i, step := range rt {
		switch step.op {
		case opUpdate:
			tr.Update(step.key, step.value)
			values[string(step.key)] = string(step.value)
		case opDelete:
			tr.Delete(step.key)
			delete(values, string(step.key))
		case opGet:
			v := tr.Get(step.key)
			want := values[string(step.key)]
			if string(v) != want {
				rt[i].err = fmt.Errorf("mismatch for key 0x%x, got 0x%x want 0x%x", step.key, v, want)
			}
		case opCommit:
			_, rt[i].err = tr.Commit(nil)
		case opHash:
			tr.Hash()
		case opReset:
			hash, err := tr.Commit(nil)
			if err != nil {
				rt[i].err = err
				return false
			}
			newtr, err := New(hash, triedb)
			if err != nil {
				rt[i].err = err
				return false
			}
			tr = newtr
		case opItercheckhash:
			checktr, _ := New(common.Hash{}, triedb)
			it := NewIterator(tr.NodeIterator(nil))
			for it.Next() {
				checktr.Update(it.Key, it.Value)
			}
			if tr.Hash() != checktr.Hash() {
				rt[i].err = fmt.Errorf("hash mismatch in opItercheckhash")
			}
		case opCheckCacheInvariant:
			rt[i].err = checkCacheInvariant(tr.root, nil, tr.cachegen, false, 0)
		}
		// Abort the test on error.
		if rt[i].err != nil {
			return false
		}
	}
	return true
}

func checkCacheInvariant(n, parent node, parentCachegen uint16, parentDirty bool, depth int) error {
	var children []node
	var flag nodeFlag
	switch n := n.(type) {
	case *shortNode:
		flag = n.flags
		children = []node{n.Val}
	case *fullNode:
		flag = n.flags
		children = n.Children[:]
	default:
		return nil
	}

	errorf := func(format string, args ...interface{}) error {
		msg := fmt.Sprintf(format, args...)
		msg += fmt.Sprintf("\nat depth %d node %s", depth, spew.Sdump(n))
		msg += fmt.Sprintf("parent: %s", spew.Sdump(parent))
		return errors.New(msg)
	}
	if flag.gen > parentCachegen {
		return errorf("cache invariant violation: %d > %d\n", flag.gen, parentCachegen)
	}
	if depth > 0 && !parentDirty && flag.dirty {
		return errorf("cache invariant violation: %d > %d\n", flag.gen, parentCachegen)
	}
	for _, child := range children {
		if err := checkCacheInvariant(child, n, flag.gen, flag.dirty, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func TestRandom(t *testing.T) {
	Convey("Test Random", t, func() {
		err := quick.Check(runRandTest, nil)
		So(err, ShouldEqual, nil)
		_, ok := err.(*quick.CheckError)
		So(false, ShouldEqual, ok)
	})
}

func BenchmarkGet(b *testing.B)      { benchGet(b, false) }
func BenchmarkGetDB(b *testing.B)    { benchGet(b, true) }
func BenchmarkUpdateBE(b *testing.B) { benchUpdate(b, binary.BigEndian) }
func BenchmarkUpdateLE(b *testing.B) { benchUpdate(b, binary.LittleEndian) }

const benchElemCount = 20000

func benchGet(b *testing.B, commit bool) {
	trie := new(Trie)
	if commit {
		_, tmpdb := tempDB()
		trie, _ = New(common.Hash{}, tmpdb)
	}
	k := make([]byte, 32)
	for i := 0; i < benchElemCount; i++ {
		binary.LittleEndian.PutUint64(k, uint64(i))
		trie.Update(k, k)
	}
	binary.LittleEndian.PutUint64(k, benchElemCount/2)
	if commit {
		trie.Commit(nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trie.Get(k)
	}
	b.StopTimer()

	if commit {
		ldb := trie.db.diskdb.(*db.LDBDatabase)
		ldb.Close()
		os.RemoveAll(ldb.Path())
	}
}

func benchUpdate(b *testing.B, e binary.ByteOrder) *Trie {
	trie := newEmpty()
	k := make([]byte, 32)
	for i := 0; i < b.N; i++ {
		e.PutUint64(k, uint64(i))
		trie.Update(k, k)
	}
	return trie
}

func BenchmarkHash(b *testing.B) {
	// Make the random benchmark deterministic
	random := rand.New(rand.NewSource(0))

	// Create a realistic account trie to hash
	addresses := make([][20]byte, b.N)
	for i := 0; i < len(addresses); i++ {
		for j := 0; j < len(addresses[i]); j++ {
			addresses[i][j] = byte(random.Intn(256))
		}
	}
	accounts := make([][]byte, len(addresses))
	for i := 0; i < len(accounts); i++ {
		var (
			nonce   = uint64(random.Int63())
			balance = new(big.Int).Rand(random, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil))
			root    = emptyRoot
			code    = Keccak256(nil)
		)
		accounts[i], _ = rlp.EncodeToBytes([]interface{}{nonce, balance, root, code})
	}
	// Insert the accounts into the trie and hash it
	trie := newEmpty()
	for i := 0; i < len(addresses); i++ {
		trie.Update(Keccak256(addresses[i][:]), accounts[i])
	}
	b.ResetTimer()
	b.ReportAllocs()
	trie.Hash()
}

func tempDB() (string, *Database) {
	dir, err := ioutil.TempDir("", "trie-bench")
	if err != nil {
		panic(fmt.Sprintf("can't create temporary directory: %v", err))
	}
	diskdb, err := db.NewLDBDatabase(dir, 256, 0)
	if err != nil {
		panic(fmt.Sprintf("can't create temporary database: %v", err))
	}
	return dir, NewDatabase(diskdb)
}

func getString(trie *Trie, k string) []byte {
	return trie.Get([]byte(k))
}

func updateString(trie *Trie, k, v string) {
	trie.Update([]byte(k), []byte(v))
}

func deleteString(trie *Trie, k string) {
	trie.Delete([]byte(k))
}
*/
