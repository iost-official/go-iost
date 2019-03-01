package database

import (
	"testing"

	"strings"

	"encoding/json"
	"os"
	"time"

	"runtime/pprof"

	"strconv"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/db"
)

func watchTimeout(t *testing.T, f func()) {
	ta := time.Now()
	f()
	tb := time.Now().Sub(ta)
	if tb > 5*time.Millisecond {
		t.Error("time overflow")
	}
}

func TestJson(t *testing.T) {
	buf, err := json.Marshal([]string{"abc", "def"})
	if err != nil {
		t.Fatal(err)
	}
	if string(buf) != `["abc","def"]` {
		t.Fatal(string(buf))
	}
}

func TestString(t *testing.T) {
	ss := strings.Replace("@a@b@c", "@b", "", 1)
	if ss != `@a@c` {
		t.Fatal(ss)
	}

	sl := strings.Split("@a@b@c", "@")
	if sl[1] != "a" {
		t.Fatal(sl)
	}
}

func TestMap(t *testing.T) {
	d, err := db.NewMVCCDB("mvcc")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		d.Close()
		os.RemoveAll("mvcc")
	}()

	f, err := os.Create("mput.prof")
	if err != nil {
		panic(err)
	}
	err = pprof.StartCPUProfile(f)
	if err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()

	vi := NewVisitor(100, d)

	watchTimeout(t, func() {
		vi.MapHandler.MPut("a", "b", "c")
	})
	watchTimeout(t, func() {
		vi.MapHandler.MPut("b", "bb", "c")
	})
	watchTimeout(t, func() {
		vi.MapHandler.MPut("c", "b", "c")
	})

	watchTimeout(t, func() {
		vi.MapHandler.MHas("a", "b")
	})
}

func TestMap303_New(t *testing.T) {
	d, err := db.NewMVCCDB("mvcc")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		d.Close()
		os.RemoveAll("mvcc")
	}()
	vi := NewVisitor(100, d)
	vi.MapHandler.db.Put("reserve.height", strconv.FormatInt(common.Ver3_0_3, 10))

	vi.MapHandler.MPut("a", "abcd", "x")
	vi.MapHandler.MPut("a", "ab", "x")
	vi.MapHandler.MPut("a", "abc", "x")
	if vi.MapHandler.db.Get("m-a") != "@@abcd@ab@abc" {
		t.Fatal(vi.MapHandler.db.Get("m-a"), "should be @@abcd@ab@abc")
	}

	vi.MapHandler.MDel("a", "ab")
	if vi.MapHandler.db.Get("m-a") != "@@abcd@abc" {
		t.Fatal(vi.MapHandler.db.Get("m-a"), "should be @@abcd@abc")
	}

	vi.MapHandler.MDel("a", "abc")
	if vi.MapHandler.db.Get("m-a") != "@@abcd" {
		t.Fatal(vi.MapHandler.db.Get("m-a"), "should be @@abcd")
	}

	vi.MapHandler.MDel("a", "abcd")
	if vi.MapHandler.db.Get("m-a") != "n" {
		t.Fatal(vi.MapHandler.db.Get("m-a"), "should be n")
	}

}
func TestMap303_ClearOld(t *testing.T) {
	d, err := db.NewMVCCDB("mvcc")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		d.Close()
		os.RemoveAll("mvcc")
	}()
	vi := NewVisitor(100, d)
	vi.MapHandler.db.Put("reserve.height", strconv.FormatInt(common.Ver3_0_3-1, 10))
	vi.MapHandler.MPut("a", "abcd", "x")
	vi.MapHandler.MPut("a", "ab", "x")
	vi.MapHandler.MPut("a", "abc", "x")
	if vi.MapHandler.db.Get("m-a") != "@abcd@ab@abc" {
		t.Fatal(vi.MapHandler.db.Get("m-a"), "should be @abcd@ab@abc")
	}

	vi.MapHandler.db.Put("reserve.height", strconv.FormatInt(common.Ver3_0_3, 10))

	vi.MapHandler.MDel("a", "abc")
	if vi.MapHandler.db.Get("m-a") != "@@abcd@ab" {
		t.Fatal(vi.MapHandler.db.Get("m-a"), "should be @@abcd@ab")
	}

}

func TestMap303_ClearOld_Err(t *testing.T) {
	d, err := db.NewMVCCDB("mvcc")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		d.Close()
		os.RemoveAll("mvcc")
	}()
	vi := NewVisitor(100, d)
	vi.MapHandler.db.Put("reserve.height", strconv.FormatInt(common.Ver3_0_3-1, 10))
	vi.MapHandler.MPut("a", "abcd", "x")
	vi.MapHandler.MPut("a", "ab", "x")
	vi.MapHandler.MPut("a", "abc", "x")
	vi.MapHandler.MDel("a", "ab")
	if vi.MapHandler.db.Get("m-a") != "cd@ab@abc" {
		t.Fatal(vi.MapHandler.db.Get("m-a"), "should be cd@ab@abc")
	}

	vi.MapHandler.db.Put("reserve.height", strconv.FormatInt(common.Ver3_0_3, 10))

	vi.MapHandler.MPut("a", "c", "x")
	if vi.MapHandler.db.Get("m-a") != "@@abc@c" {
		t.Fatal(vi.MapHandler.db.Get("m-a"), "should be @@abc@c")
	}

}
