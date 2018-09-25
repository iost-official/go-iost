package db

import (
	"math/rand"
	"os/exec"
	"testing"
	"time"
)

const (
	MaxLen = 64
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
		key := make([]byte, MaxLen)
		value := make([]byte, MaxLen)
		rand.Read(key)
		rand.Read(value)
		keys[i] = string(key)
		values[i] = string(value)
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

func BenchmarkMVCCDBPutAndCommit(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	mvccdb, err := NewMVCCDB(DBPATH)
	if err != nil {
		b.Fatalf("Failed to new mvccdb: %v", err)
	}

	keys := make([]string, b.N)
	values := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		key := make([]byte, MaxLen)
		value := make([]byte, MaxLen)
		rand.Read(key)
		rand.Read(value)
		keys[i] = string(key)
		values[i] = string(value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mvccdb.Put("table01", keys[i], values[i])
		if i%100 == 99 {
			mvccdb.Commit()
		}
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
		key := make([]byte, MaxLen)
		value := make([]byte, MaxLen)
		rand.Read(key)
		rand.Read(value)
		keys[i] = string(key)
		values[i] = string(value)
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

func BenchmarkMVCCDBGetAndCommit(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	mvccdb, err := NewMVCCDB(DBPATH)
	if err != nil {
		b.Fatalf("Failed to new mvccdb: %v", err)
	}

	keys := make([]string, b.N)
	values := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		key := make([]byte, MaxLen)
		value := make([]byte, MaxLen)
		rand.Read(key)
		rand.Read(value)
		keys[i] = string(key)
		values[i] = string(value)
	}

	for i := 0; i < b.N; i++ {
		mvccdb.Put("table01", keys[i], values[i])
		if i%100 == 99 {
			mvccdb.Commit()
		}
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
		key := make([]byte, MaxLen)
		value := make([]byte, MaxLen)
		rand.Read(key)
		rand.Read(value)
		keys[i] = string(key)
		values[i] = string(value)
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

func BenchmarkMVCCDBDelAndCommit(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	mvccdb, err := NewMVCCDB(DBPATH)
	if err != nil {
		b.Fatalf("Failed to new mvccdb: %v", err)
	}

	keys := make([]string, b.N)
	values := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		key := make([]byte, MaxLen)
		value := make([]byte, MaxLen)
		rand.Read(key)
		rand.Read(value)
		keys[i] = string(key)
		values[i] = string(value)
	}

	for i := 0; i < b.N; i++ {
		mvccdb.Put("table01", keys[i], values[i])
		if i%100 == 99 {
			mvccdb.Commit()
		}
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

func BenchmarkMVCCDBPutAndCommitAndFlush(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	mvccdb, err := NewMVCCDB(DBPATH)
	if err != nil {
		b.Fatalf("Failed to new mvccdb: %v", err)
	}

	keys := make([]string, b.N)
	values := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		key := make([]byte, MaxLen)
		value := make([]byte, MaxLen)
		rand.Read(key)
		rand.Read(value)
		keys[i] = string(key)
		values[i] = string(value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mvccdb.Put("table01", keys[i], values[i])
		if i%100 == 99 {
			mvccdb.Commit()
		}
	}
	mvccdb.Tag("blockhashcode")
	mvccdb.Flush("blockhashcode")
	b.StopTimer()

	mvccdb.Close()
	cmd := exec.Command("rm", "-r", DBPATH)
	cmd.Run()
}

func BenchmarkMVCCDBGetAndPutAndCommitAndFlush(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	mvccdb, err := NewMVCCDB(DBPATH)
	if err != nil {
		b.Fatalf("Failed to new mvccdb: %v", err)
	}

	keys := make([]string, b.N)
	values := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		key := make([]byte, MaxLen)
		value := make([]byte, MaxLen)
		rand.Read(key)
		rand.Read(value)
		keys[i] = string(key)
		values[i] = string(value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mvccdb.Get("table01", keys[i])
		mvccdb.Put("table01", keys[i], values[i])
		if i%100 == 99 {
			mvccdb.Commit()
		}
	}
	mvccdb.Tag("blockhashcode")
	mvccdb.Flush("blockhashcode")
	b.StopTimer()

	mvccdb.Close()
	cmd := exec.Command("rm", "-r", DBPATH)
	cmd.Run()
}

func BenchmarkMVCCDBCreateAndClose(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mvccdb, err := NewMVCCDB(DBPATH)
		if err != nil {
			b.Fatalf("Failed to new mvccdb: %v", err)
		}
		mvccdb.Close()
		cmd := exec.Command("rm", "-r", DBPATH)
		cmd.Run()
	}
}
