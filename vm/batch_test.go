package vm

import (
	"sync"
	"testing"
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
