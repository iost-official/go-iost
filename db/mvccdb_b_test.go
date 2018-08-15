package db

import (
	"math/rand"
	"os/exec"
	"testing"
	"time"
)

func BenchmarkMVCCDBPut(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	mvccdb, err := NewMVCCDB(DBPATH)
	if err != nil {
		b.Fatalf("Failed to new mvccdb: %v", err)
	}

	keys := make([]string, b.N)
	values := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		key := make([]byte, 32)
		value := make([]byte, 32)
		rand.Read(key)
		rand.Read(value)
		keys = append(keys, string(key))
		values = append(values, string(value))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mvccdb.Put("table01", keys[i], values[i])
	}
	b.StopTimer()

	mvccdb.Close()
	cmd := exec.Command("rm", "-r", DBPATH)
	cmd.Run()
}

func BenchmarkMVCCDBGet(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	mvccdb, err := NewMVCCDB(DBPATH)
	if err != nil {
		b.Fatalf("Failed to new mvccdb: %v", err)
	}

	keys := make([]string, b.N)
	values := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		key := make([]byte, 32)
		value := make([]byte, 32)
		rand.Read(key)
		rand.Read(value)
		keys = append(keys, string(key))
		values = append(values, string(value))
	}

	for i := 0; i < b.N; i++ {
		mvccdb.Put("table01", keys[i], values[i])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mvccdb.Get("table01", keys[i])
	}
	b.StopTimer()

	mvccdb.Close()
	cmd := exec.Command("rm", "-r", DBPATH)
	cmd.Run()
}

func BenchmarkMVCCDBDel(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	mvccdb, err := NewMVCCDB(DBPATH)
	if err != nil {
		b.Fatalf("Failed to new mvccdb: %v", err)
	}

	keys := make([]string, b.N)
	values := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		key := make([]byte, 32)
		value := make([]byte, 32)
		rand.Read(key)
		rand.Read(value)
		keys = append(keys, string(key))
		values = append(values, string(value))
	}

	for i := 0; i < b.N; i++ {
		mvccdb.Put("table01", keys[i], values[i])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mvccdb.Del("table01", keys[i])
	}
	b.StopTimer()

	mvccdb.Close()
	cmd := exec.Command("rm", "-r", DBPATH)
	cmd.Run()
}
