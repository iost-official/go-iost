package verifier

import (
	"sync"
	"testing"

	"time"

	"github.com/iost-official/go-iost/vm/database"
	"github.com/smartystreets/goconvey/convey"
)

func TestArray(t *testing.T) {
	var m = make([]int, 10)
	var w sync.WaitGroup
	for i := 0; i < 10; i++ {
		i2 := i
		w.Add(1)
		go func() {
			defer w.Done()
			time.Sleep(time.Millisecond)
			m[i2] = i2
		}()
	}
	w.Wait()
	t.Log(m)
}

func TestResolve(t *testing.T) {
	var maps = make([]map[string]database.Access, 4)
	maps[0] = map[string]database.Access{"a": 1, "b": 0}
	maps[1] = map[string]database.Access{"c": 1, "b": 0}
	maps[3] = map[string]database.Access{"b": 1, "x": 0}

	i, o := Resolve(maps)
	convey.Convey("test of resolve", t, func() {
		convey.So(i, convey.ShouldContain, 0)
		convey.So(i, convey.ShouldContain, 1)
		convey.So(o, convey.ShouldContain, 3)
	})
}

func BenchmarkResolve(b *testing.B) {
	var maps = make([]map[string]database.Access, 8)

	maps[0] = map[string]database.Access{"a": 1, "b": 0, "c": 1, "d": 1, "e": 0, "f": 1, "g": 1, "h": 0, "i": 1}
	maps[1] = map[string]database.Access{"c": 1, "d": 0, "f": 1, "c1": 1, "d1": 0, "f1": 1, "c2": 1, "d2": 0, "f2": 1}
	maps[2] = map[string]database.Access{"xx": 1, "x": 0, "y": 1, "d": 1, "e": 0, "f": 1, "g": 1, "h": 0, "i": 1}
	maps[3] = map[string]database.Access{"o": 1, "p": 0, "q": 1, "o2": 1, "p1": 0, "q1": 1, "o1": 1, "p2": 0, "q2": 1}
	maps[4] = map[string]database.Access{"j": 1, "x": 0, "y": 1, "j1": 1, "x1": 0, "y1": 1, "j2": 1, "x2": 0, "y2": 1}
	maps[5] = map[string]database.Access{"m": 1, "l": 0, "k": 1, "m1": 1, "l1": 0, "k1": 1, "m2": 1, "l2": 0, "k2": 1}
	maps[6] = map[string]database.Access{"s": 1, "g": 0, "t": 1, "s1": 1, "g1": 0, "t1": 1, "s2": 1, "g2": 0, "t2": 1}
	maps[7] = map[string]database.Access{"n": 1, "o": 0, "y": 1, "n1": 1, "o1": 0, "y1": 1, "n2": 1, "o2": 0, "y2": 1}
	for i := 0; i < b.N; i++ {
		Resolve(maps)
	}
}
