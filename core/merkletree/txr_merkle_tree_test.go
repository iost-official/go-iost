package merkletree

import (
	"testing"
	"log"
	"github.com/smartystreets/goconvey/convey"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"encoding/hex"
	"reflect"
	"bytes"
	"math/rand"
	"time"
	"fmt"
)

func TestTXRMerkleTree(t *testing.T) {
	convey.Convey("Test of TXR", t, func() {
		m := TXRMerkleTree{}
		txrs := []*tx.TxReceipt{
			{TxHash:[]byte("node1")},
			{TxHash:[]byte("node2")},
			{TxHash:[]byte("node3")},
			{TxHash:[]byte("node4")},
			{TxHash:[]byte("node5")},
		}
		err := m.Build(txrs)
		if err != nil {
			log.Panic(err)
		}
		convey.So(hex.EncodeToString(m.TX2TXR["node1"]), convey.ShouldEqual, "0a056e6f6465311a00")
		txr, err := m.GetTXR([]byte("node1"))
		if err != nil {
			log.Panic(err)
		}
		convey.So(hex.EncodeToString(txr.Encode()), convey.ShouldEqual, "0a056e6f6465311a00")

		b, err := m.Encode()
		if err != nil {
			log.Panic(err)
		}
		var m_read TXRMerkleTree
		err = m_read.Decode(b)
		if err != nil {
			log.Panic(err)
		}
		convey.So(reflect.DeepEqual(m.TX2TXR,m_read.TX2TXR), convey.ShouldBeTrue)
		for i := 0; i < 16; i++ {
			convey.So(bytes.Equal(m.MT.HashList[i], m_read.MT.HashList[i]), convey.ShouldBeTrue)
		}
	})
}

func BenchmarkTXRMerkleTree_Build(b *testing.B) { // 2439313ns = 2.4ms
	rand.Seed(time.Now().UnixNano())
	var txrs []*tx.TxReceipt
	for i := 0; i < 3000; i++ {
		fmt.Println(i)
		txrs = append(txrs, &tx.TxReceipt{TxHash:[]byte("node1")})
	}
	m := TXRMerkleTree{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := m.Build(txrs)
		if err != nil {
			log.Panic(err)
		}
	}
}