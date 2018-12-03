package wal

import (
	"encoding/binary"
	"github.com/iost-official/go-iost/db/wal/pcrc"
	"hash"
	"io"
	"os"
	"sync"
	"github.com/iost-official/go-iost/ilog"
)

// walPageBytes is the alignment for flushing logs to the backing Writer.
// It should be a multiple of the minimum sector size so that WAL can safely
// distinguish between torn writes and ordinary data corruption.
const walPageBytes = 8 * minSectorSize

type encoder struct {
	mu sync.Mutex
	w  *PageWriter

	crc       hash.Hash64
	buf       []byte
	uint64buf []byte
}

func newEncoder(w io.Writer, prevCrc uint64, pageOffset int) *encoder {
	return &encoder{
		w:   NewPageWriter(w, walPageBytes, pageOffset),
		crc: pcrc.New(prevCrc, crc64Table),
		// 1MB buffer save the allocation fee
		buf:       make([]byte, 1024*1024),
		uint64buf: make([]byte, 8),
	}
}

// newFileEncoder creates a new encoder with current file offset for the page writer.
func newFileEncoder(f *os.File, prevCrc uint64) (*encoder, error) {
	ilog.Info("Encoder Name: ", f.Name())
	offset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	return newEncoder(f, prevCrc, int(offset)), nil
}

func (e *encoder) encode(log *Log) error {
	// don't move this lock around
	// Need to ensure the crc order is same with data written order
	e.mu.Lock()
	defer e.mu.Unlock()

	e.crc.Write(log.Data)
	log.Checksum = e.crc.Sum64()
	var (
		data []byte
		err  error
		n    int
	)

	//serialize log into data
	if log.Size() > len(e.buf) {
		data, err = log.Marshal()
		if err != nil {
			return err
		}
	} else {
		n, err = log.MarshalTo(e.buf)
		if err != nil {
			return err
		}
		data = e.buf[:n]
	}

	lenField, padByteLength := encodeFrameSize(len(data))
	if err = writeUint64(e.w, lenField, e.uint64buf); err != nil {
		return err
	}

	if padByteLength != 0 {
		data = append(data, make([]byte, padByteLength)...)
	}
	_, err = e.w.Write(data)
	ilog.Info("Encoder Write: ", len(data))
	return err
}

func encodeFrameSize(dataBytes int) (lenField uint64, padByteLength int) {
	lenField = uint64(dataBytes)
	padByteLength = (8 - (dataBytes % 8)) % 8
	if padByteLength != 0 {
		//unique first bit indicate padByte existence
		lenField |= uint64(0x80|padByteLength) << 56
	}
	return lenField, padByteLength
}

func (e *encoder) flush() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.w.Flush()
}

func writeUint64(w io.Writer, n uint64, buf []byte) error {
	binary.LittleEndian.PutUint64(buf, n)
	_, err := w.Write(buf)
	return err
}
