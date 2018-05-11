package network

import (
	"testing"

	"github.com/iost-official/prototype/core/message"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNetwork(t *testing.T) {
	Convey("", t, func() {
		nn, err := NewNaiveNetwork(3)
		if err != nil {
			t.Errorf("NewNaiveNetwork encounter err %+v", err)
			return
		}
		lis1, err := nn.Listen(11037)
		So(err, ShouldBeNil)

		lis2, err := nn.Listen(11038)
		So(err, ShouldBeNil)

		req := message.Message{
			Time:    1,
			From:    "test1",
			To:      "test2",
			ReqType: 1,
			Body:    []byte{1, 1, 2},
		}
		err = nn.Broadcast(req)
		So(err, ShouldBeNil)

		message := <-lis1
		So(message.From, ShouldEqual, req.From)

		message = <-lis2
		So(message.ReqType, ShouldEqual, req.ReqType)
	})
}
