package new_vm

import (
	"testing"

	"github.com/golang/mock/gomock"
	blk "github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
)

func engineinit(t *testing.T) (*blk.BlockHead, database.IMultiValue) {
	ctl := gomock.NewController(t)
	db := database.NewMockIMultiValue(ctl)
	bh := &blk.BlockHead{
		ParentHash: []byte("abc"),
		Number:     10,
		Witness:    "witness",
		Time:       123456,
	}
	return bh, db
}

func TestNewEngine(t *testing.T) { // test of normal engine work
	bh, db := engineinit(t)
	e := NewEngine(bh, db)
	print(e.(*EngineImpl).host.ctx)
}

func TestCxt(t *testing.T) { // tests of context transport

}

func TestNative(t *testing.T) { // tests of native vm works

}
