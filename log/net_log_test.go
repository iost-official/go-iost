package log

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseMsg(t *testing.T) {
	bv := url.Values{
		"type":            []string{"Block", "sub_type"},
		"block-number":    {"2"},
		"block-head-hash": []string{"0xwwwwwwxcsdswefverr23fcdfsd"},
	}

	bt := url.Values{
		"type":      []string{"Tx", "sub_type"},
		"nonce":     []string{"2"},
		"hash":      []string{"0xwwwwwwxcsdswefverr23fcdfsd"},
		"publisher": []string{"publisher"},
	}

	bn := url.Values{
		"type": []string{"Node", "node_sub_type"},
		"log":  []string{"log"},
	}

	Convey("report test", t, func() {
		err := Report(ParseMsg(bv))
		So(err, ShouldBeNil)
		err = Report(ParseMsg(bt))
		So(err, ShouldBeNil)
		err = Report(ParseMsg(bn))
		So(err, ShouldBeNil)
	})

}
