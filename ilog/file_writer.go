package ilog

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var _ LogWriter = &FileWriter{}

type FileWriter struct {
	level    Level
	filename string
	fd       *os.File

	hourlyTicker *hourlyTicker
}

func NewFileWriter(filename string) *FileWriter {
	return &FileWriter{
		filename:     filename,
		hourlyTicker: NewHourlyTicker(),
	}
}

func (fw *FileWriter) Init() error {
	realFile := fmt.Sprintf("%s.%s", fw.filename, time.Now().Format("2006-01-02_15"))
	fd, err := os.OpenFile(realFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	fw.fd = fd
	_, err = os.Lstat(fw.filename)
	if err == nil || os.IsExist(err) {
		os.Remove(fw.filename)
	}
	os.Symlink("./"+filepath.Base(realFile), fw.filename)
	return nil
}

func (fw *FileWriter) SetLevel(l Level) {
	fw.level = l
}

func (fw *FileWriter) GetLevel() Level {
	return fw.level
}

func (fw *FileWriter) Write(msg string, level Level) error {
	fw.checkFile()
	_, err := fmt.Fprint(fw.fd, msg)
	return err
}

func (fw *FileWriter) Flush() error {
	return fw.fd.Sync()
}

func (fw *FileWriter) Close() error {
	return fw.fd.Close()
}

func (fw *FileWriter) checkFile() {
	select {
	case <-fw.hourlyTicker.C:
		fw.fd.Sync()
		fw.fd.Close()
		fw.Init()
	default:
	}
}

type hourlyTicker struct {
	C      chan time.Time
	quitCh chan struct{}
}

func NewHourlyTicker() *hourlyTicker {
	ht := &hourlyTicker{
		C:      make(chan time.Time),
		quitCh: make(chan struct{}),
	}
	ht.startTimer()
	return ht
}

func (ht *hourlyTicker) startTimer() {
	go func() {
		hour := time.Now().Hour()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ht.quitCh:
				return
			case t := <-ticker.C:
				if t.Hour() != hour {
					ht.C <- t
					hour = t.Hour()
				}
			}
		}
	}()
}

func (ht *hourlyTicker) Stop() {
	close(ht.quitCh)
}
