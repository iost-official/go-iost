package vm

import (
	"fmt"
	"sync"
	"testing"

	"github.com/iost-official/Go-IOS-Protocol/vm/database"
)

func TestArray(t *testing.T) {
	var m sync.Map
	var w sync.WaitGroup
	for i := 0; i < 10; i++ {
		i2 := i
		go func() {
			w.Add(1)
			defer w.Done()
			m.Store(i2, i2)
		}()
	}
	w.Wait()
	m.Range(func(key, value interface{}) bool {
		if key != value {
			t.Fatal(key, value)
		}
		return true
	})
}

func TestResolve(t *testing.T) {
	var maps sync.Map
	maps.Store(0, map[string]database.Access{"a": 1, "b": 0})
	maps.Store(1, map[string]database.Access{"c": 1, "d": 0})
	maps.Store(3, map[string]database.Access{"b": 1, "x": 0})

	i, o := Resolve(maps)
	fmt.Println(i, o)
}

func BenchmarkResolve(b *testing.B) {
	var maps sync.Map
	maps.Store(0, map[string]database.Access{"a": 1, "b": 0, "c": 1, "d": 1, "e": 0, "f": 1, "g": 1, "h": 0, "i": 1})
	maps.Store(1, map[string]database.Access{"c": 1, "d": 0, "f": 1, "c1": 1, "d1": 0, "f1": 1, "c2": 1, "d2": 0, "f2": 1})
	maps.Store(2, map[string]database.Access{"xx": 1, "x": 0, "y": 1, "d": 1, "e": 0, "f": 1, "g": 1, "h": 0, "i": 1})
	maps.Store(3, map[string]database.Access{"o": 1, "p": 0, "q": 1, "o2": 1, "p1": 0, "q1": 1, "o1": 1, "p2": 0, "q2": 1})
	maps.Store(4, map[string]database.Access{"j": 1, "x": 0, "y": 1, "j1": 1, "x1": 0, "y1": 1, "j2": 1, "x2": 0, "y2": 1})
	maps.Store(5, map[string]database.Access{"m": 1, "l": 0, "k": 1, "m1": 1, "l1": 0, "k1": 1, "m2": 1, "l2": 0, "k2": 1})
	maps.Store(6, map[string]database.Access{"s": 1, "g": 0, "t": 1, "s1": 1, "g1": 0, "t1": 1, "s2": 1, "g2": 0, "t2": 1})
	maps.Store(7, map[string]database.Access{"n": 1, "o": 0, "y": 1, "n1": 1, "o1": 0, "y1": 1, "n2": 1, "o2": 0, "y2": 1})
	for i := 0; i < b.N; i++ {
		Resolve(maps)
	}
}
