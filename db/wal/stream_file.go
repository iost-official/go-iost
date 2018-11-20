package wal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// StreamFile generate files like a endless stream
type StreamFile struct {
	// dir to put files
	dir string
	// size of files to make, in bytes
	sizeLimit int64
	// suffix of file to make
	suffix string
	count  int64

	fileChannel             chan *os.File
	finishedFilePathChannel chan string
	errorChannel            chan error
	finishChannel           chan struct{}
}

func newStreamFile(dir string, size int64) *StreamFile {
	return newStreamFileFull(dir, size)
}

func newStreamFileFull(dir string, size int64) *StreamFile {
	st := &StreamFile{
		dir:                     dir,
		sizeLimit:               size,
		suffix:                  "wal.tmp",
		count:                   0,
		fileChannel:             make(chan *os.File),
		errorChannel:            make(chan error),
		finishChannel:           make(chan struct{}),
		finishedFilePathChannel: make(chan string),
	}
	go st.genFile()
	return st
}

// GetNewFile get a new generated file
func (st *StreamFile) GetNewFile() (f *os.File, err error) {
	select {
	case f = <-st.fileChannel:
	case err = <-st.errorChannel:
	}
	return f, err
}

// Close close the stream file
func (st *StreamFile) Close() error {
	close(st.finishChannel)
	return <-st.errorChannel
}

// allocateFile allocate a file
func (st *StreamFile) allocateFile() (f *os.File, err error) {
	// count % 2 so this file isn't the same as the one last published
	filePath := filepath.Join(st.dir, fmt.Sprintf(".%d.%d.%s", time.Now().UnixNano(), st.count, st.suffix))
	if f, err = os.Create(filePath); err != nil {
		return nil, err
	}
	if err = f.Truncate(st.sizeLimit); err != nil {
		return nil, err
	}
	st.count = (st.count + 1) % 100
	return f, nil
}

func (st *StreamFile) genFile() {
	defer close(st.errorChannel)
	for {
		//PreAllocate a New File
		f, err := st.allocateFile()
		if err != nil {
			st.errorChannel <- err
			return
		}
		select {
		case st.fileChannel <- f:
			continue
		case <-st.finishChannel:
			os.Remove(f.Name())
			return
		}
	}
}
