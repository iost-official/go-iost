package common

//import (
//	"math"
//	"testing"
//
//	. "github.com/smartystreets/goconvey/convey"
//	"github.com/stretchr/testify/assert"
//)
//
//var intCases = []int{-1, 0, 1, math.MaxInt32}
//var byteCases = [][]byte{{255, 255, 255, 255}, {0, 0, 0, 0}, {0, 0, 0, 1}, {127, 255, 255, 255}}
//
//func TestIntToBytes(t *testing.T) {
//	for k, v := range intCases {
//		bs := IntToBytes(v)
//		assert.Equal(t, byteCases[k], bs)
//	}
//}
//
//func TestBytesToInt(t *testing.T) {
//	for k, v := range byteCases {
//		i := BytesToInt(v)
//		assert.Equal(t, intCases[k], i)
//	}
//}
//
//func TestBytesToInt64(t *testing.T) {
//	Convey("", t, func() {
//		So(BytesToInt64(Int64ToBytes(math.MaxInt64)), ShouldEqual, math.MaxInt64)
//		So(BytesToUint64(Uint64ToBytes(uint64(math.MaxUint64))), ShouldEqual, uint64(math.MaxUint64))
//	})
//}
