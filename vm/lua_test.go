package vm

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/iost-official/prototype/state/mocks"
	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/state"
)

func TestLuaVM(t *testing.T) {
	Convey("Test of Lua VM", t, func() {
		mockCtl := gomock.NewController(t)
		pool := state_mock.NewMockPool(mockCtl)

		var k state.Key

		pool.EXPECT().Put(gomock.Any(),gomock.Any()).AnyTimes().Do(func(key state.Key, value state.Value) error{
			k = key
			return nil
		})


	})
}
