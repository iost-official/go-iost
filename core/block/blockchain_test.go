package block

import (
	"crypto/rand"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewBlockChain(t *testing.T) {
	Convey("test TestNewBlockChain", t, func() {
		bc, err := NewBlockChain("./BlockChainDB/")
		So(err, ShouldBeNil)
		So(bc.Length(), ShouldEqual, bc.Length())
		fmt.Println(bc.Length())
		os.RemoveAll("./BlockChainDB/")
	})
}

func TestChainImpl(t *testing.T) {
	Convey("test Push", t, func() {
		bc, err := NewBlockChain("./BlockChainDB/")
		So(err, ShouldBeNil)
		tBlock := Block{
			Head: &BlockHead{
				Version:    2,
				ParentHash: []byte("parent Hash"),
				TxsHash:    []byte("tree hash"),
				Info:       []byte("info "),
				Number:     int64(0),
				Time:       201222,
			},
			Sign: &crypto.Signature{},
		}
		//test Push
		length := bc.Length()
		fmt.Println("length:", length)
		tBlock.Head.Number = int64(length)
		tBlock.CalculateHeadHash()
		err = bc.Push(&tBlock)
		So(err, ShouldBeNil)
		So(bc.Length(), ShouldEqual, length+1)

		//test GetBlockByNumber

		block, err := bc.GetBlockByNumber(bc.Length() - 1)
		So(err, ShouldBeNil)
		So(block, ShouldNotBeNil)
		So(block.Head.Version, ShouldEqual, 2)
		So(string(block.Head.ParentHash), ShouldEqual, string(tBlock.Head.ParentHash))
		So(string(block.Head.TxsHash), ShouldEqual, string(tBlock.Head.TxsHash))
		So(string(block.Head.Info), ShouldEqual, string(tBlock.Head.Info))
		So(block.Head.Number, ShouldEqual, tBlock.Head.Number)
		So(string(block.Head.Witness), ShouldEqual, string(tBlock.Head.Witness))
		So(string(block.Head.Time), ShouldEqual, string(tBlock.Head.Time))

		block, err = bc.Top()
		So(err, ShouldBeNil)
		So(block, ShouldNotBeNil)
		So(block.Head.Version, ShouldEqual, 2)
		So(string(block.Head.ParentHash), ShouldEqual, string(tBlock.Head.ParentHash))
		So(string(block.Head.TxsHash), ShouldEqual, string(tBlock.Head.TxsHash))
		So(string(block.Head.Info), ShouldEqual, string(tBlock.Head.Info))
		So(block.Head.Number, ShouldEqual, tBlock.Head.Number)
		So(string(block.Head.Witness), ShouldEqual, string(tBlock.Head.Witness))
		So(string(block.Head.Time), ShouldEqual, string(tBlock.Head.Time))

		HeadHash := tBlock.HeadHash()
		block, err = bc.GetBlockByHash(HeadHash)
		So(err, ShouldBeNil)
		So(block, ShouldNotBeNil)
		So(block.Head.Version, ShouldEqual, 2)
		So(string(block.Head.ParentHash), ShouldEqual, string(tBlock.Head.ParentHash))
		So(string(block.Head.TxsHash), ShouldEqual, string(tBlock.Head.TxsHash))
		So(string(block.Head.Info), ShouldEqual, string(tBlock.Head.Info))
		So(block.Head.Number, ShouldEqual, tBlock.Head.Number)
		So(string(block.Head.Witness), ShouldEqual, string(tBlock.Head.Witness))
		So(string(block.Head.Time), ShouldEqual, string(tBlock.Head.Time))
		os.RemoveAll("./BlockChainDB/")
	})
}

func BenchmarkBlock(b *testing.B) {
	a1, _ := account.NewKeyPair(nil, crypto.Secp256k1)
	a2, _ := account.NewKeyPair(nil, crypto.Secp256k1)
	actions := []*tx.Action{}
	actions = append(actions, &tx.Action{
		Contract:   "contract1",
		ActionName: "actionname1",
		Data:       "{\"num\": 1, \"message\": \"contract1\"}",
	})
	actions = append(actions, &tx.Action{
		Contract:   "contract2",
		ActionName: "actionname2",
		Data:       "1",
	})
	bc, _ := NewBlockChain("./BlockChainDB/")
	bnum := 100
	txnum := 5000
	hashes := make([][]byte, 0)
	for i := 0; i < bnum; i++ {
		parentHash := make([]byte, 32)
		rand.Read(parentHash)
		txsHash := make([]byte, 32)
		rand.Read(txsHash)
		merkleHash := make([]byte, 32)
		rand.Read(merkleHash)
		tBlock := &Block{
			Head: &BlockHead{
				Version:    2,
				ParentHash: parentHash,
				TxsHash:    txsHash,
				MerkleHash: merkleHash,
				Info:       make([]byte, 0),
				Number:     int64(i),
				Time:       time.Now().UnixNano(),
				Witness:    "IOSTfQFocqDn7VrKV7vvPqhAQGyeFU9XMYo5SNn5yQbdbzC75wM7C",
			},
		}
		for j := 0; j < txnum; j++ {
			txn := tx.NewTx(actions, []string{a1.ID, a2.ID}, 9999, 1, 1, 0)
			tBlock.Txs = append(tBlock.Txs, txn)
			tr := tx.NewTxReceipt(txn.Hash())
			tBlock.Receipts = append(tBlock.Receipts, tr)
		}
		tBlock.CalculateHeadHash()
		tBlock.Sign = a1.Sign(tBlock.HeadHash())
		hashes = append(hashes, tBlock.HeadHash())
		bc.Push(tBlock)
	}

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bc.GetBlockByHash(hashes[i%bnum])
		}
	})
	os.RemoveAll("./BlockChainDB/")
}
