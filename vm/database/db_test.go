package database

import (
	"testing"

	"os"

	. "github.com/golang/mock/gomock"
	"github.com/iost-official/go-iost/db"
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
	v.Commit()

	mockMVCC.EXPECT().Put("state", "m-hello-1", Any()).DoAndReturn(func(table string, key string, value string) error {
		if !(table == "state" && key == "m-hello-1" && value == "world") {
			t.Fatal(table, key, value)
		}
		return nil
	})

	mockMVCC.EXPECT().Put("state", "m-hello", Any()).Do(func(a, b, c string) {
		if c != "@@1" {
			t.Fatal(c)
		}
	})
	mockMVCC.EXPECT().Get("state", "reserve.height").Return("", nil)
	mockMVCC.EXPECT().Has("state", "m-hello-1").Return(false, nil)
	mockMVCC.EXPECT().Get("state", "m-hello").Return("", nil)

	v.MPut("hello", "1", "world")
	v.Commit()
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
	mockMVCC.EXPECT().Get("state", "reserve.height").Return("", nil)
	mockMVCC.EXPECT().Get("state", "m-key").Return("@a@b@c", nil)

	strs := v.MKeys("key")
	if !sliceEqual(strs, []string{"a", "b", "c"}) {
		t.Fatal(strs)
	}
}

func TestWriteCache(t *testing.T) {
	mockCtl := NewController(t)
	defer mockCtl.Finish()
	mockMVCC := NewMockIMultiValue(mockCtl)

	length := 100
	v := NewVisitor(length, mockMVCC)

	//mockMVCC.EXPECT().Put(Any(), Any(), Any()).DoAndReturn(func(table string, key string, value string) error {
	//	if !(table == "state" && key == "b-hello" && value == "world") {
	//		t.Fatal(table, key, value)
	//	}
	//	return nil
	//})
	v.Put("hello", "world")
	ok := v.Has("hello")
	if !ok {
		t.Fatal(ok)
	}
	v.Del("hello")
	ok = v.Has("hello")
	if ok {
		t.Fatal(ok)
	}
}

func TestMultiWork(t *testing.T) {
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		t.Fatal(err)
	}

	defer closeMVCCDB(mvccdb)

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

	defer closeMVCCDB(mvccdb)
	length := 100
	v1 := NewVisitor(length, mvccdb)
	v2 := NewVisitor(length, mvccdb)

	v1.Put("hello", "world")
	v1.Commit()
	vv := v2.Get("hello")
	if vv != "world" {
		t.Fatal(vv)
	}

	v2.Put("hello", "world2")
	v2.Commit()
	vv = v1.Get("hello")
	if vv != "world2" {
		t.Fatal(vv)
	}

}

func closeMVCCDB(m db.MVCCDB) {
	m.Close()
	os.RemoveAll("mvcc")
}
