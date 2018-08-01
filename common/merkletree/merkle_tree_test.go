package merkletree

import (
	"encoding/hex"
	"testing"

	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"reflect"
	"time"
)

func TestBuild(t *testing.T) {
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
	assert.Equal(
		t,
		"0f8a9f1e9450978a41ff06e77df3de64866b55261ed20651c90eb6cb462b1409",
		hex.EncodeToString(m.HashList[0]),
		"Root hash is correct",
	)
	assert.Equal(
		t,
		"e5e1a9ed8c02ed449057a4c17618127fa8e0a1e1c19fa15a371810371ac7530b",
		hex.EncodeToString(m.HashList[1]),
		"Level 1 hash 1 is correct",
	)
	assert.Equal(
		t,
		"de333248f6058db0367c9dc3e4731ea37324d4bfbbeee22ffd3d5a4e0c28330a",
		hex.EncodeToString(m.HashList[2]),
		"Level 1 hash 2 is correct",
	)
}

func TestRootHash(t *testing.T) {
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
	rootHash, err := m.RootHash()
	assert.Equal(
		t,
		"0f8a9f1e9450978a41ff06e77df3de64866b55261ed20651c90eb6cb462b1409",
		hex.EncodeToString(rootHash),
		"Root hash is correct",
	)
}

func TestMerklePath(t *testing.T) {
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
	mp, err := m.MerklePath([]byte("node5"))
	if err != nil{
		log.Panic(err)
	}
	assert.Equal(
		t,
		"6e6f646535",
		hex.EncodeToString(mp[0]),
		"path 0 is correct",
	)
	assert.Equal(
		t,
		"946f804875563d1f73fb27b1fc8af795850e9128281954028e145108db4a0ab9",
		hex.EncodeToString(mp[1]),
		"path 1 is correct",
	)
	assert.Equal(
		t,
		"e5e1a9ed8c02ed449057a4c17618127fa8e0a1e1c19fa15a371810371ac7530b",
		hex.EncodeToString(mp[2]),
		"path 1 is correct",
	)
}

func TestMerkleProve(t *testing.T) {
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
	mp, err := m.MerklePath([]byte("node5"))
	if err != nil {
		log.Panic(err)
	}
	rootHash, err := m.RootHash()
	if err != nil {
		log.Panic(err)
	}
	success, err := m.MerkleProve([]byte("node5"), rootHash, mp)
	if err != nil {
		log.Panic(err)
	}
	assert.Equal(
		t,
		true,
		success,
		"merkle prove result is correct",
	)
}

func TestSerializeAndDeserialize(t *testing.T) {
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
	var b []byte
	b, err = m.XXX_Marshal(b, false)
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
	//for i := 0; i < b.N; i++ {
	//	tmp := data[0]
	//	for j := 1; j < 1000; j++ {
	//		tmp = append(tmp, data[1]...)
	//	}
	//	sha256.Sum256(tmp)
	//}
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
