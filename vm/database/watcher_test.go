package database

import (
	"testing"

	. "github.com/golang/mock/gomock"
	"github.com/iost-official/go-iost/core/version"

	"fmt"
)

func TestWatcher(t *testing.T) {
	mockCtl := NewController(t)
	defer mockCtl.Finish()
	mockMVCC := NewMockIMultiValue(mockCtl)

	mockMVCC.EXPECT().Get("state", "b-baz").Return("hello", nil)

	bvr := NewBatchVisitorRoot(100, mockMVCC, version.NewRules(0))
	vi, watcher := NewBatchVisitor(bvr)

	vi.Put("foo", "bar")
	vi.Get("foo")

	vi.Get("baz")
	fmt.Println(watcher.Map())
}
