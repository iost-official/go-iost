package protocol_test

import (
	"reflect"
	"testing"
	"fmt"
	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/bouk/monkey"

	"github.com/iost-official/PrototypeWorks/iosbase/mocks"
	"github.com/iost-official/PrototypeWorks/iosbase"

	"github.com/iost-official/PrototypeWorks/protocol/mocks"
	. "github.com/iost-official/PrototypeWorks/protocol"
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

	Convey("Test of utils", t, func() {
		Convey("Equal of []TxInput", func() {
			a := []iosbase.TxInput{
				{
					StateHash: []byte{1, 2, 3},
				}, {
					StateHash: []byte{4, 5, 6},
				},
			}

			b := []iosbase.TxInput{
				{
					StateHash: []byte{1, 2, 3},
				}, {
					StateHash: []byte{4, 5, 6},
				},
			}

			c := []iosbase.TxInput{
				{
					StateHash: []byte{7, 8, 9},
				}, {
					StateHash: []byte{4, 5, 6},
				},
			}

			So(SliceEqualI(a, b), ShouldBeTrue)
			So(SliceEqualI(a, c), ShouldBeFalse)
		})
		Convey("Equal of []State", func() {
			a := []iosbase.State{
				{
					Value: 123,
				}, {
					Value: 456,
				},
			}

			b := []iosbase.State{
				{
					Value: 123,
				}, {
					Value: 456,
				},
			}

			c := []iosbase.State{
				{
					Value: 789,
				}, {
					Value: 456,
				},
			}

			So(SliceEqualS(a, b), ShouldBeTrue)
			So(SliceEqualS(a, c), ShouldBeFalse)
		})
		Convey("Intersect of []TxInput", func() {
			a := []iosbase.TxInput{
				{
					StateHash: []byte{1, 2, 3},
				}, {
					StateHash: []byte{4, 5, 6},
				},
			}

			b := []iosbase.TxInput{
				{
					StateHash: []byte{1, 2, 3},
				}, {
					StateHash: []byte{7, 8, 9},
				},
			}

			c := []iosbase.TxInput{
				{
					StateHash: []byte{7, 8, 9},
				}, {
					StateHash: []byte{2, 3, 1},
				},
			}

			So(SliceIntersect(a,b), ShouldBeTrue)
			So(SliceIntersect(a,c), ShouldBeFalse)
		})
		Convey("Tx conflict", func() {
			a := iosbase.Tx{
				Inputs:[]iosbase.TxInput{
					{
						StateHash: []byte{1, 2, 3},
					}, {
						StateHash: []byte{4, 5, 6},
					},
				},
				Outputs:[]iosbase.State{
					{
						Value: 123,
					}, {
						Value: 456,
					},
				},
				Recorder: "somebodyA",
			}
			b := iosbase.Tx{
				Inputs:[]iosbase.TxInput{
					{
						StateHash: []byte{1, 2, 3},
					}, {
						StateHash: []byte{4, 5, 6},
					},
				},
				Outputs:[]iosbase.State{
					{
						Value: 123,
					}, {
						Value: 456,
					},
				},
				Recorder: "somebodyB",
			}
			c := iosbase.Tx{
				Inputs:[]iosbase.TxInput{
					{
						StateHash: []byte{1, 2, 3},
					}, {
						StateHash: []byte{4, 5, 6},
					},
				},
				Outputs:[]iosbase.State{
					{
						Value: 123,
					}, {
						Value: 789,
					},
				},
				Recorder: "somebodyC",
			}
			So(TxConflict(a, b), ShouldBeTrue)
			So(TxConflict(a,c), ShouldBeFalse)
		})

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
