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
	. "github.com/smartystreets/goconvey/convey"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandHash ...
func RandHash(n int) []byte {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return []byte(string(b))
}

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
		So(hex.EncodeToString(m.HashList[0]), ShouldEqual, "1c5f24921723d90769269d6be91ae9b881b3c8e1575a8d76de5c3379c8bc14a1")
		So(hex.EncodeToString(m.HashList[1]), ShouldEqual, "1d4c19fd3644f573c1c502dd8ebb4ae1f009ccd4b21182383d4f951afbc5f0bf")
		So(hex.EncodeToString(m.HashList[2]), ShouldEqual, "ad12162552506f4c1b6075f34761130330fedbdc4eb022acb194a9c41c6aeaf4")
		rootHash := m.RootHash()
		So(hex.EncodeToString(rootHash), ShouldEqual, "1c5f24921723d90769269d6be91ae9b881b3c8e1575a8d76de5c3379c8bc14a1")
		mp, err := m.MerklePath([]byte("node5"))
		if err != nil {
			log.Panic(err)
		}
		So(hex.EncodeToString(mp[0]), ShouldEqual, "6e6f646535")
		So(hex.EncodeToString(mp[1]), ShouldEqual, "d19d621d37ab476679ff47b1e3ab8013c7bef9e7b7b18a392748fc9764d131c8")
		So(hex.EncodeToString(mp[2]), ShouldEqual, "1d4c19fd3644f573c1c502dd8ebb4ae1f009ccd4b21182383d4f951afbc5f0bf")
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
		data = append(data, RandHash(32))
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
		data = append(data, RandHash(32))
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
