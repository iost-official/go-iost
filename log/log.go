package log

import (
	"fmt"
	"time"
	"runtime/debug"
)

type Logger struct {
	Tag  string
	Path string
}

func NewLogger(tag, path string) Logger {
	l := Logger{
		Tag:  tag,
		Path: path,
	}
	return l
}

func (l *Logger) log(level, s string, attr ...interface{}) {
	a := fmt.Sprintf(s, attr...)
	str := fmt.Sprintf("%v %v/%v: %v", time.Now().Format("2006-01-02 15:04:05.000"), level, l.Tag, a)
	fmt.Println(str)
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
	fmt.Println("============CRASH")
	fmt.Printf(s , attr...)
	fmt.Println("")
	debug.PrintStack()
}
