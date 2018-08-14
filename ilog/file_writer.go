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
	filepath string
	fd       *os.File

	hourlyTicker *hourlyTicker
}

func NewFileWriter(filepath string) *FileWriter {
	return &FileWriter{
		filepath:     filepath,
		hourlyTicker: NewHourlyTicker(),
	}
}

func (fw *FileWriter) Init() error {
	if len(fw.filepath) == 0 {
		fw.filepath = "./"
	}
	if !filepath.IsAbs(fw.filepath) {
		fw.filepath, _ = filepath.Abs(fw.filepath)
	}
	if err := os.MkdirAll(fw.filepath, 0700); err != nil {
		panic(err)
	}
	realFile := fmt.Sprintf("%s/iost_%s.log", fw.filepath, time.Now().Format("2006-01-02_15"))
	linkFile := fmt.Sprintf("%s/iost.log", fw.filepath)
	fd, err := os.OpenFile(realFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	fw.fd = fd
	_, err = os.Lstat(linkFile)
	if err == nil || os.IsExist(err) {
		os.Remove(linkFile)
	}
	os.Symlink(realFile, linkFile)
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
