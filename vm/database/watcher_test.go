package database

import (
	"testing"

	. "github.com/golang/mock/gomock"

	"fmt"
)

func TestWatcher(t *testing.T) {
	mockCtl := NewController(t)
	defer mockCtl.Finish()
	mockMVCC := NewMockIMultiValue(mockCtl)

	mockMVCC.EXPECT().Get("state", "b-baz").Return("hello", nil)

	bvr := NewBatchVisitorRoot(100, mockMVCC)
	vi, watcher := NewBatchVisitor(bvr)

	vi.Put("foo", "bar")
	vi.Get("foo")

	vi.Get("baz")
	fmt.Println(watcher.Map())
}
