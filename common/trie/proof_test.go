package trie

/*
import (
	"bytes"
	crand "crypto/rand"
	mrand "math/rand"
	"testing"
	"time"

	"github.com/iost-official/prototype/db"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	mrand.Seed(time.Now().Unix())
}

func TestProof(t *testing.T) {
	Convey("Test Proof", t, func() {
		trie, vals := randomTrie(500)
		root := trie.Hash()
		for _, kv := range vals {
			proofs, _ := db.NewMemDatabase()
			So(nil, ShouldEqual, trie.Prove(kv.k, 0, proofs))
			val, err, _ := VerifyProof(root, kv.k, proofs)
			So(nil, ShouldEqual, err)
			So(true, ShouldEqual, bytes.Equal(val, kv.v))
		}
	})
}

func TestOneElementProof(t *testing.T) {
	Convey("Test OneElementProof", t, func() {
		trie := new(Trie)
		updateString(trie, "k", "v")
		proofs, _ := db.NewMemDatabase()
		trie.Prove([]byte("k"), 0, proofs)
		So(1, ShouldEqual, len(proofs.Keys()))
		val, err, _ := VerifyProof(trie.Hash(), []byte("k"), proofs)
		So(nil, ShouldEqual, err)
		So(true, ShouldEqual, bytes.Equal(val, []byte("v")))
	})
}

func TestVerifyBadProof(t *testing.T) {
	Convey("Test VerifyBadProof", t, func() {
		trie, vals := randomTrie(800)
		root := trie.Hash()
		for _, kv := range vals {
			proofs, _ := db.NewMemDatabase()
			trie.Prove(kv.k, 0, proofs)
			So(0, ShouldNotEqual, len(proofs.Keys()))
			keys := proofs.Keys()
			key := keys[mrand.Intn(len(keys))]
			node, _ := proofs.Get(key)
			proofs.Delete(key)
			mutateByte(node)
			proofs.Put(Keccak256(node), node)
			_, err, _ := VerifyProof(root, kv.k, proofs)
			So(nil, ShouldNotEqual, err)
		}
	})
}

// mutateByte changes one byte in b.
func mutateByte(b []byte) {
	for r := mrand.Intn(len(b)); ; {
		new := byte(mrand.Intn(255))
		if new != b[r] {
			b[r] = new
			break
		}
	}
}

func BenchmarkProve(b *testing.B) {
	trie, vals := randomTrie(100)
	var keys []string
	for k := range vals {
		keys = append(keys, k)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kv := vals[keys[i%len(keys)]]
		proofs, _ := db.NewMemDatabase()
		if trie.Prove(kv.k, 0, proofs); len(proofs.Keys()) == 0 {
			b.Fatalf("zero length proof for %x", kv.k)
		}
	}
}

func BenchmarkVerifyProof(b *testing.B) {
	trie, vals := randomTrie(100)
	root := trie.Hash()
	var keys []string
	var proofs []*db.MemDatabase
	for k := range vals {
		keys = append(keys, k)
		proof, _ := db.NewMemDatabase()
		trie.Prove([]byte(k), 0, proof)
		proofs = append(proofs, proof)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		im := i % len(keys)
		if _, err, _ := VerifyProof(root, []byte(keys[im]), proofs[im]); err != nil {
			b.Fatalf("key %x: %v", keys[im], err)
		}
	}
}

func randomTrie(n int) (*Trie, map[string]*kv) {
	trie := new(Trie)
	vals := make(map[string]*kv)
	for i := byte(0); i < 100; i++ {
		value := &kv{LeftPadBytes([]byte{i}, 32), []byte{i}, false}
		value2 := &kv{LeftPadBytes([]byte{i + 10}, 32), []byte{i}, false}
		trie.Update(value.k, value.v)
		trie.Update(value2.k, value2.v)
		vals[string(value.k)] = value
		vals[string(value2.k)] = value2
	}
	for i := 0; i < n; i++ {
		value := &kv{randBytes(32), randBytes(20), false}
		trie.Update(value.k, value.v)
		vals[string(value.k)] = value
	}
	return trie, vals
}

func randBytes(n int) []byte {
	r := make([]byte, n)
	crand.Read(r)
	return r
}
*/
