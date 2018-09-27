package merkletree

import (
	"testing"

	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMerkleTree(t *testing.T) {
	Convey("Test of MT", t, func() {
		data := [][]byte{
			[]byte("node1"),
			[]byte("node2"),
			[]byte("node3"),
			[]byte("node4"),
			[]byte("node5"),
		}
		m := MerkleTree{}
		m.Build(data)
		So(hex.EncodeToString(m.HashList[0]), ShouldEqual, "0f8a9f1e9450978a41ff06e77df3de64866b55261ed20651c90eb6cb462b1409")
		So(hex.EncodeToString(m.HashList[1]), ShouldEqual, "e5e1a9ed8c02ed449057a4c17618127fa8e0a1e1c19fa15a371810371ac7530b")
		So(hex.EncodeToString(m.HashList[2]), ShouldEqual, "de333248f6058db0367c9dc3e4731ea37324d4bfbbeee22ffd3d5a4e0c28330a")
		rootHash := m.RootHash()
		So(hex.EncodeToString(rootHash), ShouldEqual, "0f8a9f1e9450978a41ff06e77df3de64866b55261ed20651c90eb6cb462b1409")
		mp, err := m.MerklePath([]byte("node5"))
		if err != nil {
			log.Panic(err)
		}
		So(hex.EncodeToString(mp[0]), ShouldEqual, "6e6f646535")
		So(hex.EncodeToString(mp[1]), ShouldEqual, "946f804875563d1f73fb27b1fc8af795850e9128281954028e145108db4a0ab9")
		So(hex.EncodeToString(mp[2]), ShouldEqual, "e5e1a9ed8c02ed449057a4c17618127fa8e0a1e1c19fa15a371810371ac7530b")
		success, _ := m.MerkleProve([]byte("node5"), rootHash, mp)
		So(success, ShouldBeTrue)
		b, err := proto.Marshal(&m)
		if err != nil {
			log.Panic(err)
		}
		var mRead MerkleTree
		err = proto.Unmarshal(b, &mRead)
		if err != nil {
			log.Panic(err)
		}
		for i := 0; i < 16; i++ {
			So(bytes.Equal(m.HashList[i], mRead.HashList[i]), ShouldBeTrue)
		}
	})
}

func BenchmarkBuild(b *testing.B) { // 646503ns = 0.6ms，vs 117729ns = 0.1ms
	rand.Seed(time.Now().UnixNano())
	var data [][]byte
	for i := 0; i < 2; i++ {
		fmt.Println(i)
		data = append(data, common.RandHash(32))
	}
	m := MerkleTree{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Build(data)
	}
}

func BenchmarkMerklePath(b *testing.B) { // 183ns
	rand.Seed(time.Now().UnixNano())
	var data [][]byte
	for i := 0; i < 1000; i++ {
		data = append(data, common.RandHash(32))
	}
	m := MerkleTree{}
	m.Build(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		datum := data[rand.Intn(1000)]
		_, err := m.MerklePath(datum)
		if err != nil {
			log.Panic(err)
		}
	}
}
