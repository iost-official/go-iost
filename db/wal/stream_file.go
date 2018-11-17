package wal

import (
	"os"
	"path/filepath"
	"fmt"
	"time"
)

const (
	defaultFlushInterval = 50 * time.Millisecond // defaultFlushInterval Set to 50 Millisecond
	defaultFlushSize = 4096 // defaultFlushInterval Set to 50 Millisecond
)

type StreamFile struct {
	// dir to put files
	dir string
	// size of files to make, in bytes
	sizeLimit int64
	// suffix of file to make
	suffix string
	count int64

	fileChannel   chan *os.File
	finishedFilePathChannel chan string
	errorChannel  chan error
	finishChannel chan struct{}
}

func newStreamFile(dir string, size int64) *StreamFile {
	return newStreamFileFull(dir, size)
}

func newStreamFileFull(dir string, size int64) *StreamFile {
	st := &StreamFile{
		dir:           dir,
		sizeLimit:          size,
		suffix:		   "wal.tmp",
		count: 0,
		fileChannel:   make(chan *os.File),
		errorChannel:  make(chan error),
		finishChannel: make(chan struct{}),
		finishedFilePathChannel: make(chan string),
	}
	go st.GenFile()
	return st
}

func (st * StreamFile) GetNewFile() (f *os.File, err error){
	select {
		case f =<- st.fileChannel:
		case err =<- st.errorChannel:
	}
	return f, err
}

func (st * StreamFile) Close() error{
	close(st.finishChannel)
	return <- st.errorChannel
}

func (st * StreamFile) AllocateFile() (f *os.File, err error){
	// count % 2 so this file isn't the same as the one last published
	filePath := filepath.Join(st.dir, fmt.Sprintf(".%d.%d.%s", time.Now().UnixNano(), st.count, st.suffix))
	if f, err = os.Create(filePath); err != nil {
		return nil, err
	}
	if err = f.Truncate(st.sizeLimit); err != nil {
		return nil, err
	}
	st.count = (st.count+1) % 100
	return f, nil
}

func (st * StreamFile) GenFile(){
	defer close(st.errorChannel)
	for {
		//PreAllocate a New File
		f, err := st.AllocateFile()
		if err != nil{
			st.errorChannel<-err
			return
		}
		select {
		case st.fileChannel <-f:
			continue
		case <-st.finishChannel:
			os.Remove(f.Name())
			return
		}
	}
}
