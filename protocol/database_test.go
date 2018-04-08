package protocol_test

import (
	"fmt"
	. "github.com/bouk/monkey"
	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"

	"github.com/iost-official/Go-IOS-Protocol/iosbase"
	"github.com/iost-official/Go-IOS-Protocol/iosbase/mocks"

	. "github.com/iost-official/Go-IOS-Protocol/protocol"
	"github.com/iost-official/Go-IOS-Protocol/protocol/mocks"
)

func TestDatabase(t *testing.T) {
	Convey("Test of Database factory：", t, func() {
		mock := NewController(t)
		defer mock.Finish()

		db, err := DatabaseFactory(
			"base",
			iosbase_mock.NewMockBlockChain(mock),
			iosbase_mock.NewMockStatePool(mock))

		So(err, ShouldBeNil)
		So(reflect.TypeOf(db), ShouldEqual, reflect.TypeOf(&DatabaseImpl{}))
	})

	Convey("Test of DatabaseImpl", t, func() {
		mock := NewController(t)
		defer mock.Finish()

		bc := iosbase_mock.NewMockBlockChain(mock)
		sp := iosbase_mock.NewMockStatePool(mock)
		tp := iosbase_mock.NewMockTxPool(mock)

		db, _ := DatabaseFactory("base", bc, sp)

		Convey("Verify Tx", func() {
			tx1 := iosbase.Tx{
				Inputs: []iosbase.TxInput{
					{
						TxHash:    []byte{1, 2, 3},
						StateHash: []byte{4, 5, 6},
					},
				},
			}
			tx2 := iosbase.Tx{
				Inputs: []iosbase.TxInput{
					{
						TxHash:    []byte{1, 2, 3},
						StateHash: []byte{7, 8, 9},
					},
				},
			}

			sp.EXPECT().Find([]byte{4, 5, 6}).AnyTimes().Return(iosbase.State{}, nil)
			sp.EXPECT().Find([]byte{7, 8, 9}).AnyTimes().Return(iosbase.State{}, fmt.Errorf("error"))

			err := db.VerifyTx(tx1)
			So(err, ShouldBeNil)
			err = db.VerifyTx(tx2)
			So(err, ShouldNotBeNil)
		})

		Convey("Verify Tx with cache", func() {
			tx0 := iosbase.Tx{
				Inputs: []iosbase.TxInput{
					{
						StateHash: []byte{1, 2, 3},
					},
				},
			}
			tx1 := iosbase.Tx{
				Inputs: []iosbase.TxInput{
					{
						StateHash: []byte{4, 5, 6},
					},
				},
			}
			tx2 := iosbase.Tx{
				Inputs: []iosbase.TxInput{
					{
						StateHash: []byte{7, 8, 9},
					},
				},
			}
			sp.EXPECT().Find([]byte{1, 2, 3}).AnyTimes().Return(iosbase.State{}, nil)
			sp.EXPECT().Find([]byte{4, 5, 6}).AnyTimes().Return(iosbase.State{}, nil)
			sp.EXPECT().Find([]byte{7, 8, 9}).AnyTimes().Return(iosbase.State{}, fmt.Errorf("error"))

			tp.EXPECT().GetSlice().AnyTimes().Return([]iosbase.Tx{tx1}, nil)

			err := db.VerifyTxWithCache(tx0, tp)
			So(err, ShouldBeNil)
			err = db.VerifyTxWithCache(tx1, tp)
			So(err, ShouldNotBeNil)
			err = db.VerifyTxWithCache(tx2, tp)
			So(err, ShouldNotBeNil)
		})

		Convey("Push Block and get new view", func() {
			guard := Patch(ViewFactory, func(target string) (View, error) {
				view := protocol_mock.NewMockView(mock)
				view.EXPECT().Init(Any()).AnyTimes()
				view.EXPECT().ByzantineTolerance().AnyTimes().Return(5)
				return view, nil
			})
			defer guard.Unpatch()

			bc.EXPECT().Push(Any()).AnyTimes()
			sp.EXPECT().Transact(Any()).AnyTimes()

			chv1, err := db.NewViewSignal()
			chv2, err := db.NewViewSignal()

			So(err, ShouldBeNil)

			var v1, v2 View
			go func() {
				v1 = <-chv1
			}()

			go func() {
				v2 = <-chv2
			}()

			db.PushBlock(&iosbase.Block{})

			So(v1.ByzantineTolerance(), ShouldEqual, 5)
			So(v2.ByzantineTolerance(), ShouldEqual, 5)

		})

	})

}
