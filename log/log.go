package log

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

type Logger struct {
	Tag     string
	logFile *os.File
}

var instance *Logger

var initialized int32
var mu sync.Mutex

func GetLogger(tag, path string) (*Logger, error) {

	if atomic.LoadInt32(&initialized) == 1 {
		return instance, nil
	}

	mu.Lock()
	defer mu.Unlock()

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	instance = &Logger{
		Tag:     tag,
		logFile: file,
	}

	runtime.SetFinalizer(instance, func(obj *Logger) {
		obj.logFile.Close()
	})

	atomic.StoreInt32(&initialized, 1)

	return instance, nil

}

func (l *Logger) log(level, s string, attr ...interface{}) {
	a := fmt.Sprintf(s, attr...)
	str := fmt.Sprintf("%v %v/%v: %v", time.Now().Format("2006-01-02 15:04:05.000"), level, l.Tag, a)
	l.logFile.Write([]byte(str))
	l.logFile.Write([]byte("\n"))
}

func (l *Logger) D(s string, attr ...interface{}) {
	l.log("D", s, attr...)
}

func (l *Logger) I(s string, attr ...interface{}) {
	l.log("I", s, attr...)
}

func (l *Logger) E(s string, attr ...interface{}) {
	l.log("E", s, attr...)
}

func (l *Logger) Crash(s string, attr ...interface{}) {
	l.logFile.Write([]byte("============CRASH\n"))
	fs := fmt.Sprintf(s, attr...)
	l.logFile.Write([]byte(fs))
	l.logFile.Write([]byte("\n"))
	l.logFile.Write(debug.Stack())
}
