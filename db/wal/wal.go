// Package wal This Module is in so many aspects inspired by etcd's WAL.
package wal

import (
	"bytes"
	"errors"
	"fmt"
	"hash/crc64"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/ilog"
)

var (
	// SegmentSizeBytes is the preallocated size of each wal segment file.
	// The actual size might be larger than this. In general, the default
	// value should be used, but this is defined as an exported variable
	// so that tests can set a different segment size.
	SegmentSizeBytes int64 = 8 * 1000 * 1000 // 8MB

	// ErrMetadataConflict metadata not consist
	ErrMetadataConflict = errors.New("wal: conflicting metadata found")
	// ErrFileNotFound file not found
	ErrFileNotFound = errors.New("wal: file not found")
	// ErrCRCMismatch crc miss match
	ErrCRCMismatch   = errors.New("wal: crc mismatch")
	crc64Table       = crc64.MakeTable(crc64.ECMA)
	warnSyncDuration = time.Second
)

// WAL is a logical representation of the stable storage.
// WAL is either in read mode or append mode but not both.
// A newly created WAL is in append mode, and ready for appending records.
// A just opened WAL is in read mode, and ready for reading records.
// The WAL will be ready for appending after reading out all the previous records.
type WAL struct {
	dir string // the living directory of the underlay files

	// dirFile is a fd for the wal directory for syncing on Rename
	dirFile *os.File

	metadata []byte // metadata recorded at the head of each WAL
	//state    raftpb.HardState // hardstate recorded at the head of WAL

	//start     walpb.Snapshot // snapshot to start reading
	decoder   *decoder     // decoder to decode records
	readClose func() error // closer for decode reader

	mu             sync.Mutex
	lastEntryIndex uint64   // index of the last entry saved to the wal
	encoder        *encoder // encoder to encode records

	files []*os.File // the locked files the WAL holds (the name is increasing)
	st    *StreamFile
}

// Create creates a WAL ready for appending records. The given metadata is
// recorded at the head of each WAL file, and can be retrieved with ReadAll.
func Create(dirpath string, metadata []byte) (*WAL, error) {
	b, err := exists(dirpath)
	if err != nil {
		return nil, err
	}
	if !b {
		err := os.MkdirAll(dirpath, 0777)
		if err != nil {
			ilog.Error("failed to create a WAL Directory! error: ", err)
			return nil, err
		}
	}

	if Exist(dirpath) {
		// Recover
		return recoverFromDir(dirpath, metadata)
	}

	streamFile := newStreamFile(dirpath, SegmentSizeBytes)
	f, err := streamFile.GetNewFile()
	if err != nil {
		ilog.Warn("failed to generate a new WAL temp file！", err)
		return nil, err
	}

	w := &WAL{
		dir:      dirpath,
		metadata: metadata,
		st:       streamFile,
	}

	if w.dirFile, err = OpenDir(w.dir); err != nil {
		return w, err
	}

	w.encoder, err = newFileEncoder(f, 0)
	if err != nil {
		return nil, err
	}
	w.files = append(w.files, f)
	if err = w.saveCrc(0); err != nil {
		return nil, err
	}
	if err = w.encoder.encode(&Log{Type: LogType_metaDataType, Data: metadata}); err != nil {
		return nil, err
	}
	w.encoder.flush()

	return w, nil
}

func recoverFromDir(dirpath string, metadata []byte) (*WAL, error) {
	ilog.Info("RecoverFromDir")
	w, err := Open(dirpath)
	if err != nil {
		return nil, err
	}

	return w, err

}

// Open opens the WAL at the given snap.
// The snap SHOULD have been previously saved to the WAL, or the following
// ReadAll will fail.
// The returned WAL is ready to read and the first record will be the one after
// the given snap. The WAL cannot be appended to before reading out all of its
// previous records.
func Open(dirpath string) (*WAL, error) {
	w, err := openAtIndex(dirpath)
	if err != nil {
		return nil, err
	}
	if w.dirFile, err = OpenDir(w.dir); err != nil {
		return nil, err
	}
	return w, nil
}

// OpenForRead only opens the wal files for read.
// Write on a read only wal panics.
func OpenForRead(dirpath string) (*WAL, error) {
	return openAtIndex(dirpath)
}

func openAtIndex(dirpath string) (*WAL, error) {
	names, err := readWALNames(dirpath)
	if err != nil {
		return nil, err
	}

	// open the wal files
	rcs := make([]io.ReadCloser, 0)
	rs := make([]io.Reader, 0)
	ls := make([]*os.File, 0)
	for _, name := range names {
		p := filepath.Join(dirpath, name)
		l, err := os.OpenFile(p, os.O_RDWR, 0666)
		if err != nil {
			closeAll(rcs...)
			return nil, err
		}
		ls = append(ls, l)
		rs = append(rs, l)
		if strings.HasSuffix(name, ".wal") {
			rcs = append(rcs, l)
		}
	}

	closer := func() error { return closeAll(rcs...) }

	streamFile := newStreamFile(dirpath, SegmentSizeBytes)
	// create a WAL ready for reading
	w := &WAL{
		dir:       dirpath,
		decoder:   newDecoder(rs...),
		readClose: closer,
		files:     ls,
		st:        streamFile,
	}

	return w, nil
}

// nolint
// ReadAll reads out records of the current WAL.
// After ReadAll, the WAL will be ready for appending new records.
func (w *WAL) ReadAll() (metadata []byte, ents []Entry, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	log := &Log{}
	if w.decoder == nil {
		return nil, nil, errors.New("Wal Has No Decoder!")

	}
	decoder := w.decoder

	for err = decoder.decode(log); err == nil; err = decoder.decode(log) {
		switch log.Type {
		case LogType_entryType:
			e := mustUnmarshalEntry(log.Data)
			ents = append(ents, e)
			w.lastEntryIndex = e.Index

		case LogType_metaDataType:
			if metadata != nil && !bytes.Equal(metadata, log.Data) {
				return nil, nil, ErrMetadataConflict
			}
			metadata = log.Data

		case LogType_crcType:
			crc := decoder.crc.Sum64()
			// current crc of decoder must match the crc of the record.
			// do no need to match 0 crc, since the decoder is a new one at this case.
			if crc != 0 && log.Check(crc) != nil {
				return nil, nil, ErrCRCMismatch
			}
			decoder.updateCRC(log.Checksum)

		default:
			return nil, nil, fmt.Errorf("unexpected block type %d", log.Type)
		}
	}

	switch w.tail() {
	case nil:
		// We do not have to read out all entries in read mode.
		// The last record maybe a partial written one, so
		// ErrunexpectedEOF might be returned.
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			return nil, nil, err
		}
	default:
		// We must read all of the entries if WAL is opened in write mode.
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			return nil, nil, err
		}
		// decodeRecord() will return io.EOF if it detects a zero record,
		// but this zero record may be followed by non-zero records from
		// a torn write. Overwriting some of these non-zero records, but
		// not all, will cause CRC errors on WAL open. Since the records
		// were never fully synced to disk in the first place, it's safe
		// to zero them out to avoid any CRC errors from new writes.
		if _, err = w.tail().Seek(w.decoder.getLastOffset(), io.SeekStart); err != nil {
			return nil, nil, err
		}
		if err = ZeroToEnd(w.tail()); err != nil {
			return nil, nil, err
		}
	}

	err = nil

	// close decoder, disable reading
	if w.readClose != nil {
		w.readClose()
		w.readClose = nil
	}

	w.metadata = metadata

	if w.tail() != nil {
		if !strings.HasSuffix(w.tail().Name(), ".wal.tmp") {
			f, error := w.st.GetNewFile()
			if error != nil {
				err = error
				return
			}
			w.files = append(w.files, f)
		}
		// create encoder (chain crc with the decoder), enable appending
		w.encoder, err = newFileEncoder(w.tail(), w.decoder.lastCRC())
		if err != nil {
			return
		}
	}
	w.decoder = nil

	return metadata, ents, err
}

// RemoveFiles remove files less than index
func (w *WAL) RemoveFiles(index uint64) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	fileIndex := -1
	for i, file := range w.files {
		fileName := file.Name()
		if strings.HasSuffix(fileName, ".wal.tmp") {
			continue
		}
		_, lastIndex, err := parseWALName(fileName)
		if err != nil {
			return err
		}
		if lastIndex <= index {
			if i == len(w.files)-1 {
				continue
			}
			fileIndex = i
			file.Close()
			err = os.Remove(fileName)
			if err != nil {
				return err
			}
			continue
		}
	}
	w.files = w.files[fileIndex+1:]
	return nil
}

// cut closes current file written and creates a new one ready to append.
// cut first creates a temp wal file and writes necessary headers into it.
// Then cut atomically rename temp wal file to a wal file.
func (w *WAL) cut() error {
	// close old wal file; truncate to avoid wasting space if an early cut
	off, serr := w.tail().Seek(0, io.SeekCurrent)
	if serr != nil {
		return serr
	}

	if err := w.tail().Truncate(off); err != nil {
		return err
	}

	if err := w.sync(); err != nil {
		return err
	}

	fpath := filepath.Join(w.dir, walName(w.seq()+1, w.lastEntryIndex))

	if err := os.Rename(w.tail().Name(), fpath); err != nil {
		return err
	}

	var err error
	w.files[len(w.files)-1], err = os.Open(fpath)
	if err != nil {
		return err
	}

	// create a temp wal file with name sequence + 1, or truncate the existing one
	newTail, err := w.st.GetNewFile()
	if err != nil {
		return err
	}

	// update writer and save the previous crc
	w.files = append(w.files, newTail)
	prevCrc := w.encoder.crc.Sum64()
	w.encoder, err = newFileEncoder(w.tail(), prevCrc)
	if err != nil {
		return err
	}

	if err = w.saveCrc(prevCrc); err != nil {
		return err
	}

	if err = w.encoder.encode(&Log{Type: LogType_metaDataType, Data: w.metadata}); err != nil {
		return err
	}
	// atomically move temp wal file to wal file
	if err = w.sync(); err != nil {
		return err
	}

	_, err = w.tail().Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	ilog.Info("created a new WAL segment, old tail file moved to: ", fpath)
	return nil
}

func (w *WAL) sync() error {
	if w.encoder != nil {
		if err := w.encoder.flush(); err != nil {
			return err
		}
	}
	start := time.Now()

	err := w.tail().Sync()
	took := time.Since(start)
	if took > warnSyncDuration {
		ilog.Warn("slow fdatasync, took", took, " , expected-duration", warnSyncDuration)
	}
	//walFsyncSec.Observe(took.Seconds())

	return err
}

// ReleaseLockTo releases the files, which has smaller index than the given index
// except the largest one among them.
// For example, if WAL is holding lock 1,2,3,4,5,6, ReleaseLockTo(4) will release
// lock 1,2 but keep 3. ReleaseLockTo(5) will release 1,2,3 but keep 4.
func (w *WAL) ReleaseLockTo(index uint64) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.files) == 0 {
		return nil
	}

	var smaller int
	found := false
	for i, l := range w.files {
		_, lockIndex, err := parseWALName(filepath.Base(l.Name()))
		if err != nil {
			return err
		}
		if lockIndex >= index {
			smaller = i - 1
			found = true
			break
		}
	}

	// if no lock index is greater than the release index, we can
	// release lock up to the last one(excluding).
	if !found {
		smaller = len(w.files) - 1
	}

	if smaller <= 0 {
		return nil
	}

	for i := 0; i < smaller; i++ {
		if w.files[i] == nil {
			continue
		}
		w.files[i].Close()
	}
	w.files = w.files[smaller:]

	return nil
}

// Close closes the current WAL file and directory.
func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.st != nil {
		w.st.Close()
		w.st = nil
	}

	if w.tail() != nil {
		if err := w.sync(); err != nil {
			return err
		}
	}
	for _, l := range w.files {
		if l == nil {
			continue
		}
		if err := l.Close(); err != nil {
			ilog.Warn("failed to close WAL, error: ", err)
		}
	}

	return w.dirFile.Close()
}

func (w *WAL) saveEntry(e *Entry) error {
	// TODO: add MustMarshalTo to reduce one allocation.
	e.Index = w.lastEntryIndex
	w.lastEntryIndex++
	b, err := proto.Marshal(e)
	if err != nil {
		return err
	}

	log := &Log{Type: LogType_entryType, Data: b}
	if err := w.encoder.encode(log); err != nil {
		return err
	}
	//w.lastEntryIndex = e.Index
	return nil
}

// SaveSingle save single entry
func (w *WAL) SaveSingle(ent Entry) (uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// TODO(xiangli): no more reference operator
	if err := w.saveEntry(&ent); err != nil {
		return 0, err
	}

	if err := w.encoder.flush(); err != nil {
		return 0, err
	}

	curOff, err := w.tail().Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	if curOff < SegmentSizeBytes {
		return w.lastEntryIndex, nil
	}

	return w.lastEntryIndex, w.cut()
}

// Save save entries
func (w *WAL) Save(ents []Entry) (uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(ents) == 0 {
		return w.lastEntryIndex, nil
	}

	// TODO(xiangli): no more reference operator
	for i := range ents {
		if err := w.saveEntry(&ents[i]); err != nil {
			return 0, err
		}
	}

	curOff, err := w.tail().Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	if curOff < SegmentSizeBytes {
		return w.lastEntryIndex, nil
	}

	return w.lastEntryIndex, w.cut()
}

// HasDecoder check whether wal has decoder
func (w *WAL) HasDecoder() bool {
	if w.decoder != nil && len(w.decoder.r) != 0 {
		return true
	}
	return false

}

// called when new file is used.
func (w *WAL) saveCrc(prevCrc uint64) error {
	return w.encoder.encode(&Log{Type: LogType_crcType, Data: Uint64ToBytes(prevCrc)})
}

func (w *WAL) tail() *os.File {
	if len(w.files) > 0 {
		return w.files[len(w.files)-1]
	}
	return nil
}

func (w *WAL) seq() uint64 {
	if len(w.files) <= 1 {
		return 0
	}
	t := w.files[len(w.files)-2]
	if t == nil {
		return 0
	}
	seq, _, err := parseWALName(filepath.Base(t.Name()))
	if err != nil {
		ilog.Fatal("failed to parse WAL name, name: ", t.Name(), " , error: ", err)
	}
	return seq
}

func closeAll(rcs ...io.ReadCloser) error {
	for _, f := range rcs {
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}

// CleanDir clean the wal dir
func (w *WAL) CleanDir() error {
	if w.dirFile != nil {
		return os.RemoveAll(w.dirFile.Name())
	}
	return nil
}

// Size return WAL used data size include current tmp file.
func (w *WAL) Size() uint64 {
	size := uint64(len(w.files)) * uint64(SegmentSizeBytes)
	return size
}
