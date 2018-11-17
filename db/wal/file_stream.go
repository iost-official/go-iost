package wal

import (
	"os"
	"path/filepath"
	"fmt"
	"time"
)

type FileStream struct {
	// dir to put files
	dir string
	// size of files to make, in bytes
	sizeLimit int64
	// suffix of file to make
	suffix string
	count int64

	fileChannel   chan *os.File
	errorChannel  chan error
	finishChannel chan struct{}
}


func newFileStream(dir string, size int64, suffix string) *FileStream {
	st := &FileStream{
		dir:           dir,
		sizeLimit:          size,
		suffix:		   suffix,
		count: 0,
		fileChannel:   make(chan *os.File),
		errorChannel:  make(chan error),
		finishChannel: make(chan struct{}),
	}
	go st.GenFile()
	return st
}

func (st * FileStream) GetNewFile() (f *os.File, err error){
	select {
	case f =<- st.fileChannel:
	case err =<- st.errorChannel:
	}
	return f, err
}

func (st * FileStream) Close() error{
	close(st.finishChannel)
	return <- st.errorChannel
}

func (st * FileStream) AllocateFile() (f *os.File, err error){
	// count % 2 so this file isn't the same as the one last published
	filePath := filepath.Join(st.dir, fmt.Sprintf("%d.%d.%s.tmp", time.Nanosecond, st.count, st.suffix))
	//CreateFile
	if f, err = os.Create(filePath); err != nil {
		return nil, err
	}
	//Truncate File
	if err = f.Truncate(st.sizeLimit); err != nil {
		return nil, err
	}
	st.count = (st.count+1) % 100
	return f, nil

}

func (st * FileStream) GenFile(){
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
			return
		}
	}
}
