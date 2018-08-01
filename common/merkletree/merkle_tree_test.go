package merkletree

import (
	"testing"

	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"math/rand"
	"time"
	"encoding/hex"
	"reflect"
	"github.com/stretchr/testify/assert"
	"bytes"
	"github.com/golang/protobuf/proto"

)

func TestSerializeAndDeserialize(t *testing.T) {
	data := [][]byte{
		[]byte("node1"),
		[]byte("node2"),
		[]byte("node3"),
		[]byte("node4"),
		[]byte("node5"),
	}
	m := MerkleTree{}
	m.Build(data)
	var b []byte
	b, err := m.XXX_Marshal(b, false)
	if err != nil {
		log.Panic(err)
	}
	m_read := MerkleTree{}
	err = m_read.XXX_Unmarshal(b)
	if err != nil {
		log.Panic(err)
	}
	assert.Equal(
		t,
		true,
		reflect.DeepEqual(m.Hash2Idx, m_read.Hash2Idx),
		"Hash2Idx is equal",
	)
	assert.Equal(
		t,
		true,
		reflect.DeepEqual(m.HashList, m_read.HashList),
		"HashList is equal",
	)
}

func TestMerkleTree(t *testing.T) {
	Convey("Test of Build", t, func() {
		data := [][]byte{
			[]byte("node1"),
			[]byte("node2"),
			[]byte("node3"),
			[]byte("node4"),
			[]byte("node5"),
		}
		m := MerkleTree{}
		err := m.Build(data)
		if err != nil {
			log.Panic(err)
		}
		So(hex.EncodeToString(m.HashList[0]), ShouldEqual, "0f8a9f1e9450978a41ff06e77df3de64866b55261ed20651c90eb6cb462b1409")
		So(hex.EncodeToString(m.HashList[1]), ShouldEqual, "e5e1a9ed8c02ed449057a4c17618127fa8e0a1e1c19fa15a371810371ac7530b")
		So(hex.EncodeToString(m.HashList[2]), ShouldEqual, "de333248f6058db0367c9dc3e4731ea37324d4bfbbeee22ffd3d5a4e0c28330a")
		rootHash, err := m.RootHash()
		So(hex.EncodeToString(rootHash), ShouldEqual, "0f8a9f1e9450978a41ff06e77df3de64866b55261ed20651c90eb6cb462b1409")
		mp, err := m.MerklePath([]byte("node5"))
		if err != nil {
			log.Panic(err)
		}
		So(hex.EncodeToString(mp[0]), ShouldEqual, "6e6f646535")
		So(hex.EncodeToString(mp[1]), ShouldEqual, "946f804875563d1f73fb27b1fc8af795850e9128281954028e145108db4a0ab9")
		So(hex.EncodeToString(mp[2]), ShouldEqual, "e5e1a9ed8c02ed449057a4c17618127fa8e0a1e1c19fa15a371810371ac7530b")
		success, err := m.MerkleProve([]byte("node5"), rootHash, mp)
		So(success, ShouldBeTrue)
		b, err := proto.Marshal(&m)
		if err != nil {
			log.Panic(err)
		}
		var m_read MerkleTree
		err = proto.Unmarshal(b, &m_read)
		if err != nil {
			log.Panic(err)
		}
		for i := 0; i < 16; i++ {
			So(bytes.Equal(m.HashList[i], m_read.HashList[i]), ShouldBeTrue)
		}
	})
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func BenchmarkBuild(b *testing.B) { // 646503ns = 0.6ms，vs 117729ns = 0.1ms
	rand.Seed(time.Now().UnixNano())
	var data [][]byte
	for i := 0; i < 2; i++ {
		fmt.Println(i)
		data = append(data, []byte(RandStringRunes(32)))
	}
	m := MerkleTree{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := m.Build(data)
		if err != nil {
			log.Panic(err)
		}
	}
}

func BenchmarkMerklePath(b *testing.B) { // 183ns
	rand.Seed(time.Now().UnixNano())
	var data [][]byte
	for i := 0; i < 1000; i++ {
		data = append(data, []byte(RandStringRunes(32)))
	}
	m := MerkleTree{}
	err := m.Build(data)
	if err != nil {
		log.Panic(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		datum := data[rand.Intn(1000)]
		_, err := m.MerklePath(datum)
		if err != nil {
			log.Panic(err)
		}
	}
}
