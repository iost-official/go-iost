package merkletree

import (
	"bytes"
	"log"
	"math/rand"
	"os/exec"
	"reflect"
	"testing"
	"time"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/smartystreets/goconvey/convey"
)

func TestTXRMerkleTreeDB(t *testing.T) {
	convey.Convey("Test of TXRMTDB", t, func() {
		m := TXRMerkleTree{}
		txrs := []*tx.TxReceipt{
			tx.NewTxReceipt([]byte("node1")),
			tx.NewTxReceipt([]byte("node2")),
			tx.NewTxReceipt([]byte("node3")),
			tx.NewTxReceipt([]byte("node4")),
			tx.NewTxReceipt([]byte("node5")),
		}
		m.Build(txrs)
		Init("./")
		err := TXRMTDB.Put(&m, 32342)
		if err != nil {
			log.Panic(err)
		}
		var mRead *TXRMerkleTree
		mRead, err = TXRMTDB.Get(32342)
		if err != nil {
			log.Panic(err)
		}
		convey.So(reflect.DeepEqual(m.Tx2Txr, mRead.Tx2Txr), convey.ShouldBeTrue)
		for i := 0; i < 16; i++ {
			convey.So(bytes.Equal(m.Mt.HashList[i], mRead.Mt.HashList[i]), convey.ShouldBeTrue)
		}
		cmd := exec.Command("rm", "-r", "./TXRMerkleTreeDB")
		cmd.Run()
	})
}

func BenchmarkTXRMerkleTreeDB(b *testing.B) { //Put: 1544788ns = 1.5ms, Get: 621922ns = 0.6ms
	rand.Seed(time.Now().UnixNano())
	Init("./")
	var txrs []*tx.TxReceipt
	for i := 0; i < 3000; i++ {
		txrs = append(txrs, tx.NewTxReceipt([]byte("node1")))
	}
	m := TXRMerkleTree{}
	m.Build(txrs)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TXRMTDB.Put(&m, 33)
		TXRMTDB.Get(33)
	}
}
