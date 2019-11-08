package ilog

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var _ LogWriter = &FileWriter{}

// FileWriter is responsible for writing log to file.
type FileWriter struct {
	level    Level
	filepath string
	fd       *os.File

	hourlyTicker *hourlyTicker
}

// NewFileWriter returns a new instance of NewFileWriter.
func NewFileWriter(filepath string) *FileWriter {
	return &FileWriter{
		filepath:     filepath,
		hourlyTicker: newHourlyTicker(),
	}
}

// Init inits NewFileWriter.
func (fw *FileWriter) Init() error {
	if len(fw.filepath) == 0 {
		fw.filepath = "./"
	}
	if !filepath.IsAbs(fw.filepath) {
		fw.filepath, _ = filepath.Abs(fw.filepath)
	}
	if err := os.MkdirAll(fw.filepath, 0755); err != nil {
		panic(err)
	}
	logFile := fmt.Sprintf("iost_%s.log", time.Now().Format("2006-01-02_15"))
	linkFile := filepath.Join(fw.filepath, "iost.log")

	_, err := os.Lstat(linkFile)
	if err == nil || os.IsExist(err) {
		os.Remove(linkFile)
	}
	os.Symlink(logFile, linkFile)

	fd, err := os.OpenFile(filepath.Join(fw.filepath, logFile), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	fw.fd = fd
	return nil
}

// SetLevel sets the log level.
func (fw *FileWriter) SetLevel(l Level) {
	fw.level = l
}

// GetLevel returns the log level.
func (fw *FileWriter) GetLevel() Level {
	return fw.level
}

// Write writes message to the console.
func (fw *FileWriter) Write(msg string, level Level) error {
	fw.checkFile()
	_, err := fmt.Fprint(fw.fd, msg)
	return err
}

// Flush flushes.
func (fw *FileWriter) Flush() error {
	return fw.fd.Sync()
}

// Close closes the writer.
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

func newHourlyTicker() *hourlyTicker {
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

func (ht *hourlyTicker) stop() { // nolint:unused
	close(ht.quitCh)
}
