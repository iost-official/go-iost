package wal

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"math"
	"testing"
	"strings"
)

func TestNew(t *testing.T) {
	p, err := ioutil.TempDir(os.TempDir(), "waltest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(p)

	w, err := Create(p, []byte("somedata"))
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if g := filepath.Base(w.tail().Name()); !strings.HasSuffix(g, ".wal.tmp") {
		t.Errorf("tmp file not end with .wal.tmp, has name: %s", g)
	}
	defer w.Close()

	// file is preallocated to segment size; only read data written by wal
	off, err := w.tail().Seek(0, io.SeekCurrent)
	if err != nil {
		t.Fatal(err)
	}
	gd := make([]byte, off)
	f, err := os.Open(filepath.Join(p, filepath.Base(w.tail().Name())))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err = io.ReadFull(f, gd); err != nil {
		t.Fatalf("err = %v, want nil", err)
	}

	var wb bytes.Buffer
	e := newEncoder(&wb, 0, 0)
	err = e.encode(&Log{Type: LogType_crcType, Data: Uint64ToBytes(0)})
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	err = e.encode(&Log{Type: LogType_metaDataType, Data: []byte("somedata")})
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}

	e.flush()
	if !bytes.Equal(gd, wb.Bytes()) {
		t.Errorf("data = %v, want %v", gd, wb.Bytes())
	}
}

func TestCreateFailFromNoSpaceLeft(t *testing.T) {
	p, err := ioutil.TempDir(os.TempDir(), "waltest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(p)

	oldSegmentSizeBytes := SegmentSizeBytes
	defer func() {
		SegmentSizeBytes = oldSegmentSizeBytes
	}()
	SegmentSizeBytes = math.MaxInt64

	_, err = Create(p, []byte("data"))
	if err == nil { // no space left on device
		t.Fatalf("expected error 'no space left on device', got nil")
	}
}

func TestSave(t *testing.T) {
	p, err := ioutil.TempDir(os.TempDir(), "waltest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(p)

	w, err := Create(p, []byte("somedata"))
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if g := filepath.Base(w.tail().Name()); !strings.HasSuffix(g, ".wal.tmp") {
		t.Errorf("tmp file not end with .wal.tmp, has name: %s", g)
	}

	ents := make([]Entry, 0)
	entry1 := Entry{
		Data: []byte("Entry1"),
	}
	entry2 := Entry{
		Data: []byte("Entry2"),
	}
	ents = append(ents, entry1)
	ents = append(ents, entry2)
	w.Save(ents)
	w.Save(ents)
	w.Close()
	newW, err := Create(p, []byte("somedata"))
	if len(newW.decoder.r) == 0 {
		t.Fatal("Decoder has no reader!")
	}
	metad, entries, err := newW.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if string(metad) != "somedata"{
		t.Fatal("metadata not consistent! Got: ", string(metad), " expect: somedata")
	}

	if len(entries) != 4 {
		t.Fatal("Entry length not match, should be 4, got: ", len(entries))
	}

	if entries[0].Index != 0{
		t.Fatal("Entry Index miss match, should be 0, got: ", entries[0].Index)
	}
	if entries[1].Index != 1{
		t.Fatal("Entry Index miss match, should be 1, got: ", entries[0].Index)
	}
	if entries[2].Index != 2{
		t.Fatal("Entry Index miss match, should be 2, got: ", entries[0].Index)
	}
	if entries[3].Index != 3{
		t.Fatal("Entry Index miss match, should be 3, got: ", entries[0].Index)
	}
}

func TestSaveWithCut(t *testing.T) {
	p, err := ioutil.TempDir(os.TempDir(), "waltest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(p)

	w, err := Create(p, []byte("somedata"))
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if g := filepath.Base(w.tail().Name()); !strings.HasSuffix(g, ".wal.tmp") {
		t.Errorf("tmp file not end with .wal.tmp, has name: %s", g)
	}

	ents := make([]Entry, 0)
	entry1 := Entry{
		Data: []byte("Entry1"),
	}
	entry2 := Entry{
		Data: []byte("Entry2"),
	}
	ents = append(ents, entry1)
	ents = append(ents, entry2)
	w.Save(ents)
	w.cut()
	w.Save(ents)
	w.Close()
	newW, err := Create(p, []byte("somedata"))
	if len(newW.decoder.r) == 0 {
		t.Fatal("Decoder has no reader!")
	}
	metad, entries, err := newW.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if string(metad) != "somedata"{
		t.Fatal("metadata not consistent! Got: ", string(metad), " expect: somedata")
	}

	if len(entries) != 4 {
		t.Fatal("Entry length not match, should be 4, got: ", len(entries))
	}

	if entries[0].Index != 0{
		t.Fatal("Entry Index miss match, should be 0, got: ", entries[0].Index)
	}
	if entries[1].Index != 1{
		t.Fatal("Entry Index miss match, should be 1, got: ", entries[0].Index)
	}
	if entries[2].Index != 2{
		t.Fatal("Entry Index miss match, should be 2, got: ", entries[0].Index)
	}
	if entries[3].Index != 3{
		t.Fatal("Entry Index miss match, should be 3, got: ", entries[0].Index)
	}
}
