package common

import (
	"math"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

var intCases = []int{-1, 0, 1, math.MaxInt32}
var byteCases = [][]byte{{255, 255, 255, 255}, {0, 0, 0, 0}, {0, 0, 0, 1}, {127, 255, 255, 255}}

func TestIntToBytes(t *testing.T) {
	for k, v := range intCases {
		bs := IntToBytes(v)
		assert.Equal(t, byteCases[k], bs)
	}
}

func TestBytesToInt(t *testing.T) {
	for k, v := range byteCases {
		i := BytesToInt(v)
		assert.Equal(t, intCases[k], i)
	}
}

func TestBytesToInt64(t *testing.T) {
	Convey("", t, func() {
		var mySlice = []byte{21, 43, 115, 131, 2, 137, 44, 146}
		So(Int64ToBytes(BytesToInt64(mySlice)), ShouldEqual, mySlice)
	})
}
