package wal

import (
	"bufio"
	"encoding/binary"
	"hash"
	"io"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/v3/db/wal/pcrc"
	"github.com/iost-official/go-iost/v3/ilog"
)

const (
	minSectorSize = 512
)

type decoder struct {
	mu         sync.Mutex
	r          []*bufio.Reader
	crc        hash.Hash64
	lastOffset int64
}

const frameSizeLength = 8 // record current frame size in a int64 which is 8 bytes

func newDecoder(r ...io.Reader) *decoder {
	readers := make([]*bufio.Reader, len(r))
	for i := range r {
		readers[i] = bufio.NewReader(r[i])
	}
	return &decoder{
		r:   readers,
		crc: pcrc.New(0, crc64Table),
	}
}

func (d *decoder) decode(log *Log) error {
	log.Reset()
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.decodeRecord(log)
}

func (d *decoder) decodeRecord(log *Log) error {
	if len(d.r) == 0 {
		return io.EOF
	}

	l, err := readInt64(d.r[0])
	if err == io.EOF || (err == nil && l == 0) {
		// hit end of file or preallocated space
		d.r = d.r[1:]
		if len(d.r) == 0 {
			return io.EOF
		}
		d.lastOffset = 0
		return d.decodeRecord(log)
	}
	if err != nil {
		return err
	}

	// Now we know how many bytes we used for this record.
	recBytes, padBytes := decodeFrameSize(l)

	data := make([]byte, recBytes+padBytes)
	if _, err = io.ReadFull(d.r[0], data); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}
	//Decode the record.
	if err := proto.Unmarshal(data[:recBytes], log); err != nil {
		if d.isTornWrite(data) {
			return io.ErrUnexpectedEOF
		}
		ilog.Error("Failed to unmarshal Log: ", data)
		return err
	}

	// skip pcrc checking if the record type is crcType
	if log.Type != LogType_crcType {
		d.crc.Write(log.Data)
		if err := log.Check(d.crc.Sum64()); err != nil {
			if d.isTornWrite(data) {
				return io.ErrUnexpectedEOF
			}
			return err
		}
	}
	// Got a record, update last offset
	d.lastOffset += frameSizeLength + recBytes + padBytes
	return nil
}

func decodeFrameSize(lenField int64) (recBytes int64, padBytes int64) {
	// the record size is stored in the lower 56 bits of the 64-bit length
	recBytes = int64(uint64(lenField) & ^(uint64(0xff) << 56))
	// non-zero padding is indicated by set MSb / a negative length
	if lenField < 0 { // padding is stored in lower 3 bits of length MSB
		padBytes = int64((uint64(lenField) >> 56) & 0x7)
	}
	return recBytes, padBytes
}

// isTornWrite: check is data intact, last log written may not be torn write
func (d *decoder) isTornWrite(data []byte) bool {
	//if there are more files, this one must be intact
	if len(d.r) != 1 {
		return false
	}

	fileOff := d.lastOffset + frameSizeLength
	curOff := 0
	chunks := [][]byte{}
	// split data on sector boundaries
	for curOff < len(data) {
		// the first Chunk might not be a full Sector
		currentChunkLen := int(minSectorSize - (fileOff % minSectorSize))

		// the last Chunk might not be a full Sector
		if currentChunkLen > len(data)-curOff {
			currentChunkLen = len(data) - curOff
		}
		chunks = append(chunks, data[curOff:curOff+currentChunkLen])
		fileOff += int64(currentChunkLen)
		curOff += currentChunkLen
	}

	// if any data for a sector chunk is all 0, it's a torn write
	for _, sect := range chunks {
		isZero := true
		for _, v := range sect {
			if v != 0 {
				isZero = false
				break
			}
		}
		if isZero {
			return true
		}
	}
	return false
}

func (d *decoder) updateCRC(prevCrc uint64) {
	d.crc = pcrc.New(prevCrc, crc64Table)
}

func (d *decoder) lastCRC() uint64 {
	return d.crc.Sum64()
}

func (d *decoder) getLastOffset() int64 { return d.lastOffset }

func mustUnmarshalEntry(d []byte) *Entry {
	var e Entry
	if err := proto.Unmarshal(d, &e); err != nil {
		ilog.Fatal("unmarshal should never fai")
	}
	return &e
}

func readInt64(r io.Reader) (int64, error) {
	var n int64
	err := binary.Read(r, binary.LittleEndian, &n)
	return n, err
}
