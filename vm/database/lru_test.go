package database

import (
	"testing"

	lru "github.com/hashicorp/golang-lru/v2"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLRU(t *testing.T) {
	lru0, err := lru.New[int, int](5)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		lru0.Add(i, i)
		lru0.Get(3)
	}

	Convey("test of lru", t, func() {
		So(lru0.Len(), ShouldEqual, 5)
		So(lru0.Contains(3), ShouldBeTrue)
		So(lru0.Contains(5), ShouldBeFalse)
	})
}
