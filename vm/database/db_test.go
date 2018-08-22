package database

import (
	"testing"

	. "github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"os"
)

func sliceEqual(a, b []string) bool {
	if len(a) == len(b) {
		for i, s := range a {
			if s != b[i] {
				return false
			}
		}
		return true
	}
	return false
}

func TestHandler_Put(t *testing.T) {
	mockCtl := NewController(t)
	defer mockCtl.Finish()
	mockMVCC := NewMockIMultiValue(mockCtl)

	length := 100
	v := NewVisitor(length, mockMVCC)

	mockMVCC.EXPECT().Put(Any(), Any(), Any()).DoAndReturn(func(table string, key string, value string) error {
		if !(table == "state" && key == "b-hello" && value == "world") {
			t.Fatal(table, key, value)
		}
		return nil
	})
	v.Put("hello", "world")

	mockMVCC.EXPECT().Put(Any(), Any(), Any()).DoAndReturn(func(table string, key string, value string) error {
		if !(table == "state" && key == "m-hello-1" && value == "world") {
			t.Fatal(table, key, value)
		}
		return nil
	})
	v.MPut("hello", "1", "world")
}

func TestHandler_Get(t *testing.T) {
	mockCtl := NewController(t)
	defer mockCtl.Finish()
	mockMVCC := NewMockIMultiValue(mockCtl)

	length := 100
	v := NewVisitor(length, mockMVCC)

	// test of Get
	mockMVCC.EXPECT().Get(Any(), Any()).DoAndReturn(func(table string, key string) (value string, err error) {
		if !(table == "state" && key == "b-hello") {
			t.Fatal(table, key)
		}
		return "world", nil
	})
	vv := v.Get("hello")
	if !(vv == "world") {
		t.Fatal(vv)
	}

	// test of MGet
	mockMVCC.EXPECT().Get(Any(), Any()).DoAndReturn(func(table string, key string) (value string, err error) {
		if !(table == "state" && key == "m-hello-1") {
			t.Fatal(table, key, value)
		}
		return "world", nil
	})
	vv = v.MGet("hello", "1")
	if !(vv == "world") {
		t.Fatal(vv)
	}

	// test of MKeys
	mockMVCC.EXPECT().Keys(Any(), Any()).DoAndReturn(func(table string, prefix string) ([]string, error) {
		if !(table == "state" && prefix == "m-key-") {
			t.Fatal(table, prefix)
		}
		return []string{"m-key-a", "m-key-b", "m-key-c"}, nil
	})

	strs := v.MKeys("key")
	if !sliceEqual(strs, []string{"a", "b", "c"}) {
		t.Fatal(strs)
	}
}

func TestMultiWork(t *testing.T) {
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll("mvcc")

	length := 100

	v := NewVisitor(length, mvccdb)

	v.Put("hello", "world")

	vv := v.Get("hello")
	if !(vv == "world") {
		t.Fatal(vv)
	}

	v.Put("hello", "world2")

	vv = v.Get("hello")
	if !(vv == "world2") {
		t.Fatal(vv)
	}
}

func TestMultiVisitor(t *testing.T) {
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll("mvcc")
	length := 100
	v1 := NewVisitor(length, mvccdb)
	v2 := NewVisitor(length, mvccdb)

	v1.Put("hello", "world")
	vv := v2.Get("hello")
	if vv != "world" {
		t.Fatal(vv)
	}

	v2.Put("hello", "world2")
	vv = v1.Get("hello")
	if vv != "world2" {
		t.Fatal(vv)
	}

}
