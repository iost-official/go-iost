package protocol

import (
	"reflect"
	"testing"

	. "github.com/bouk/monkey"
	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/iost-official/Go-IOS-Protocol/iosbase"
	. "github.com/iost-official/Go-IOS-Protocol/iosbase/mocks"
)

func TestReplica_Unit(t *testing.T) {
	mock := NewController(t)
	defer mock.Finish()

	var guards []*PatchGuard
	defer func() {
		for _, g := range guards {
			g.Unpatch()
		}
	}()

	bc := NewMockBlockChain(mock)
	bc.EXPECT().Top().AnyTimes().Return(&iosbase.Block{
		Head: iosbase.BlockHead{
			Version: 122,
		},
	})

	tp := NewMockTxPool(mock)
	tp.EXPECT().Hash().AnyTimes().Return([]byte{1, 2, 3})
	tp.EXPECT().Encode().AnyTimes().Return([]byte{1, 2, 3})

	v := PobView{}
	guards = append(guards, PatchInstanceMethod(
		reflect.TypeOf(&v),
		"IsPrimary",
		func(a *PobView, id string) bool {
			if id == "a" {
				return true
			}
			return false
		}))
	guards = append(guards, PatchInstanceMethod(
		reflect.TypeOf(&v),
		"IsBackup",
		func(a *PobView, id string) bool {
			if id == "b" {
				return true
			}
			return false
		}))

	db := DatabaseImpl{}
	guards = append(guards, PatchInstanceMethod(
		reflect.TypeOf(&db),
		"GetBlockChain",
		func(d *DatabaseImpl) (iosbase.BlockChain, error) {
			return bc, nil
		}))

	router := RouterImpl{}
	guards = append(guards, PatchInstanceMethod(
		reflect.TypeOf(&router),
		"Send",
		func(a *RouterImpl, req iosbase.Request) chan iosbase.Response {
			return nil
		},
	))

	mem, _ := iosbase.NewMember(nil)

	rep := ReplicaImpl{
		Member: mem,
		view:   &v,
		db:     &db,
		net:    &router,
		txPool: tp,
		block: &iosbase.Block{
			Head: iosbase.BlockHead{
				Time: 12345,
			},
		},
	}

	Convey("Test of utils", t, func() {

		Convey("Make new block:", func() {
			blk, err := rep.makeBlock()
			So(err, ShouldBeNil)
			So(blk.Head.TreeHash[0], ShouldEqual, 1)
		})
		Convey("Make prepare vote", func() {
			p := rep.makePrepare(true)
			So(iosbase.VerifySignature(p.Rand, p.Pubkey, p.Sig), ShouldBeFalse)
			So(p.Rand[0], ShouldEqual, 0x00)
		})
		Convey("Make Commit", func() {
			c := rep.makeCommit()
			So(iosbase.VerifySignature(c.BlkHeadHash, c.Pubkey, c.Sig), ShouldBeTrue)
		})
	})

	Convey("Test of PBFT phases", t, func() {
		Convey("On start phase", func() {
			var p Phase
			rep.Member.ID = "a"
			p, _ = rep.onStart()
			So(rep.chara, ShouldEqual, Primary)
			So(p, ShouldEqual, PreparePhase)

			rep.Member.ID = "b"
			p, _ = rep.onStart()
			So(rep.chara, ShouldEqual, Backup)
			So(p, ShouldEqual, PrePreparePhase)

			rep.Member.ID = "c"
			p, _ = rep.onStart()
			So(rep.chara, ShouldEqual, Idle)
			So(p, ShouldEqual, StartPhase)
		})
	})
}
