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
	Tag       string
	logFile   *os.File
	NeedPrint bool
}

func init() {
	var err error
	Log, err = NewLogger("iost")
	if err != nil {
		panic(err)
	}
}

// global Logger instance
var Log *Logger
var instance *os.File

var initialized int32
var mu sync.Mutex

const Path = "test.log" // TODO : 在命令行内修改

func NewLogger(tag string) (*Logger, error) {

	if atomic.LoadInt32(&initialized) == 1 {
		Log = &Logger{
			logFile: instance,
		}
		return Log, nil
	}

	mu.Lock()
	defer mu.Unlock()

	file, err := os.OpenFile(Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	instance = file

	runtime.SetFinalizer(instance, func(obj *os.File) {
		obj.Close()
	})

	atomic.StoreInt32(&initialized, 1)
	Log = &Logger{
		logFile:   instance,
		NeedPrint: false,
	}
	return Log, nil

}

func (l *Logger) log(level, s string, attr ...interface{}) {
	a := fmt.Sprintf(s, attr...)
	str := fmt.Sprintf("%v %v/%v: %v", time.Now().Format("2006-01-02 15:04:05.000"), level, l.Tag, a)
	if l.NeedPrint {
		fmt.Println(str)
	}
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
