package verifier

import (
	"testing"

	. "github.com/golang/mock/gomock"
	"github.com/iost-official/go-iost/core/block"
	vmMock "github.com/iost-official/go-iost/vm/mock"

	"fmt"
	"time"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/vm"
	"github.com/iost-official/go-iost/vm/database"
)

func initProvider(controller *Controller) vm.Provider {
	provider := vmMock.NewMockProvider(controller)

	txs := []*tx.Tx{
		{Time: 0},
		{Time: 1},
		{Time: 2},
		{Time: 3},
		{Time: 4},
	}

	index := 0

	provider.EXPECT().Tx().AnyTimes().DoAndReturn(func() *tx.Tx {
		i := index
		index++
		return txs[i]
	})
	return nil
}

func TestBase(t *testing.T) {
	t.Skip()
	mc := NewController(t)
	e := vmMock.NewMockEngine(mc)
	provider := initProvider(mc)
	db := database.NewMockIMultiValue(mc)

	var blk block.Block
	t1 := time.Now()
	err := baseGen(&blk, db, provider, e, &Config{
		Mode:        0,
		Timeout:     time.Second,
		TxTimeLimit: time.Millisecond,
	})
	t2 := time.Now().Sub(t1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(t2)
}
