package protocol

import (
	"reflect"
	"testing"

	. "github.com/bouk/monkey"
	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/iost-official/PrototypeWorks/iosbase"
	. "github.com/iost-official/PrototypeWorks/iosbase/mocks"
)

func TestReplica_Unit(t *testing.T) {
	Convey("Test of utils", t, func() {

	})

	Convey("Test of PBFT phases", t, func() {
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

		v := DposView{}
		guards = append(guards, PatchInstanceMethod(
			reflect.TypeOf(&v),
			"IsPrimary",
			func(a *DposView, id string) bool {
				if id == "a" {
					return true
				}
				return false
			}))
		guards = append(guards, PatchInstanceMethod(
			reflect.TypeOf(&v),
			"IsBackup",
			func(a *DposView, id string) bool {
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

		Convey("On start phase", func() {

			rep := ReplicaImpl{
				//Member: iosbase.Member{ID: "a"},
				view: &v,
				db:   &db,
				net:  &router,
			}
			var p Phase

			rep.Member.ID = "a"
			p, _ = rep.onStart()
			So(rep.chara, ShouldEqual, Primary)
			So(p, ShouldEqual, StartPhase)

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
