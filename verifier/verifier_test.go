package verifier

import (
	"testing"
	"time"

	"os"

	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/db"
)

type mockTxIter struct {
}

func (m *mockTxIter) Next() (*tx.Tx, bool) {
	return nil, false
}

func TestVerifier_Gen(t *testing.T) {
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("mvcc")

	mti := mockTxIter{}

	blk := block.Block{
		Head: &block.BlockHead{
			Version:    0,
			ParentHash: []byte{},
			Number:     0,
			Witness:    "abc",
			Time:       time.Now().UnixNano(),
			GasUsage:   0,
		},
		Txs:      []*tx.Tx{},
		Receipts: []*tx.TxReceipt{},
	}

	var v Verifier
	a, b, c := v.Gen(&blk, nil, mvccdb, &mti, &Config{
		Mode:        0,
		Timeout:     time.Second,
		TxTimeLimit: time.Millisecond * 100,
	})

	t.Log(a, b, c)
	t.Log(blk)

	err = v.Verify(&blk, nil, mvccdb, &Config{
		Mode:        0,
		Timeout:     time.Second,
		TxTimeLimit: time.Millisecond * 100,
	})

	if err != nil {
		t.Fatal(err)
	}
}
