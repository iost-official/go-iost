package log

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

type Logger struct {
	Tag          string
	logFile      *os.File
	logFileStart time.Time
	NeedPrint    bool
}

func init() {
	Log = &Logger{
		Tag:          "init",
		logFile:      newLogFile(time.Now()),
		NeedPrint:    false,
		logFileStart: time.Now(),
	}
}

// global Logger instance
var Log *Logger

const (
	ExpireTime = 24 * 60 * 60
	RotateTime = 15 * 60
)

func NewLogger(tag string) (*Logger, error) {
	Log.Tag = tag
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
	l.handleFiles()
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

const FmtTime = "2006-01-02_15:04:05.000"

func (l *Logger) handleFiles() {
	now := time.Now()
	if now.Unix()-l.logFileStart.Unix() > RotateTime {
		clearOldLogs(now)
		l.logFile = newLogFile(now)
		l.logFileStart = now
	}
}

func newLogFile(now time.Time) *os.File {
	path := now.Format(FmtTime) + ".log"
	os.Mkdir("logs/", 0777)
	file, err := os.OpenFile("logs/"+path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	return file
}

func clearOldLogs(now time.Time) {
	files, _ := ioutil.ReadDir("logs/")
	for _, f := range files {
		name := f.Name()
		if strings.HasSuffix(name, ".log") {
			timestamp := name[:strings.LastIndex(name, ".")]
			//fmt.Println(timestamp)
			mtime, err := time.Parse(FmtTime, timestamp)
			if err != nil {
				fmt.Println(err)
			}
			ts := now.Format(FmtTime)
			now2, _ := time.Parse(FmtTime, ts)
			if now2.Unix()-mtime.Unix() > ExpireTime {
				err := os.Remove("logs/" + f.Name())
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}

func ofTime() {
	now := time.Now()
	ts := now.Format(FmtTime)
	now2, _ := time.Parse(FmtTime, ts)
	fmt.Println(now.Unix(), now2.Unix())
}
