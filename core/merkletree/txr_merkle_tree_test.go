package merkletree

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"os"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	"github.com/smartystreets/goconvey/convey"
)

func TestTXRMerkleTree(t *testing.T) {
	convey.Convey("Test of TXR", t, func() {
		m := TXRMerkleTree{}
		txrs := []*tx.TxReceipt{
			tx.NewTxReceipt([]byte("node1")),
			tx.NewTxReceipt([]byte("node2")),
			tx.NewTxReceipt([]byte("node3")),
			tx.NewTxReceipt([]byte("node4")),
			tx.NewTxReceipt([]byte("node5")),
		}
		m.Build(txrs)
		convey.So(hex.EncodeToString(m.Tx2Txr["node1"]), convey.ShouldEqual, "966ca83b0c876b72584a2c1fd8757c9191d83b40953e3f939c0b1afb22da37c3")
		txr, err := m.GetTXR([]byte("node1"))
		ilog.Info(err)
		if err != nil {
			log.Panic(err)
		}
		convey.So(hex.EncodeToString(txr.Encode()), convey.ShouldEqual, "0a056e6f6465312200")

		b, err := m.Encode()
		if err != nil {
			log.Panic(err)
		}
		var mRead TXRMerkleTree
		err = mRead.Decode(b)
		if err != nil {
			log.Panic(err)
		}
		convey.So(reflect.DeepEqual(m.Tx2Txr, mRead.Tx2Txr), convey.ShouldBeTrue)
		for i := 0; i < 16; i++ {
			convey.So(bytes.Equal(m.Mt.HashList[i], mRead.Mt.HashList[i]), convey.ShouldBeTrue)
		}
	})
}

func BenchmarkTXRMerkleTree_Build(b *testing.B) { // 2439313ns = 2.4ms
	rand.Seed(time.Now().UnixNano())
	var txrs []*tx.TxReceipt
	for i := 0; i < 3000; i++ {
		fmt.Println(i)
		txrs = append(txrs, tx.NewTxReceipt([]byte("node1")))
	}
	m := TXRMerkleTree{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Build(txrs)
	}
	os.RemoveAll("TXRMerkleTreeDB")
}
