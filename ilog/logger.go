package ilog

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type message struct {
	content string
	level   Level
}

type Logger struct {
	callDepth   int
	lowestLevel Level
	isRunning   int32

	writers []LogWriter
	msg     chan *message

	bufPool *BufPool

	quitCh chan struct{}
}

func New() *Logger {
	return &Logger{
		callDepth:   2,
		lowestLevel: LevelFatal,
		msg:         make(chan *message, 1024),
		bufPool:     NewBufPool(),
		quitCh:      make(chan struct{}),
	}
}

func NewConsoleLogger() *Logger {
	logger := New()
	consoleWriter := NewConsoleWriter()
	logger.AddWriter(consoleWriter)
	return logger
}

func (logger *Logger) Start() {
	if !atomic.CompareAndSwapInt32(&logger.isRunning, 0, 1) {
		return
	}
	if len(logger.writers) == 0 {
		fmt.Println("logger's writers is empty.")
		return
	}

	go func() {
		for {
			select {
			case <-logger.quitCh:
				return
			case msg := <-logger.msg:
				logger.Write(msg)
			}
		}
	}()
}

func (logger *Logger) Stop() {
	if !atomic.CompareAndSwapInt32(&logger.isRunning, 1, 0) {
		return
	}
	close(logger.quitCh)
	logger.cleanMsg()
	logger.Flush()
}

func (logger *Logger) SetCallDepth(d int) {
	logger.callDepth = d
}

func (logger *Logger) AddWriter(writer LogWriter) error {
	if err := writer.Init(); err != nil {
		return err
	}
	logger.writers = append(logger.writers, writer)
	if logger.lowestLevel > writer.GetLevel() {
		logger.lowestLevel = writer.GetLevel()
	}
	return nil
}

func (logger *Logger) Write(msg *message) {
	wg := &sync.WaitGroup{}
	for _, writer := range logger.writers {
		if msg.level < writer.GetLevel() {
			continue
		}
		wg.Add(1)
		go func(lw LogWriter) {
			lw.Write(msg.content, msg.level)
			wg.Done()
		}(writer)
	}
	wg.Wait()

}

func (logger *Logger) Flush() {
	wg := &sync.WaitGroup{}
	for _, writer := range logger.writers {
		wg.Add(1)
		go func(lw LogWriter) {
			lw.Flush()
			wg.Done()
		}(writer)
	}
	wg.Wait()
}

func (logger *Logger) Debug(format string, v ...interface{}) {
	logger.genMsg(LevelDebug, fmt.Sprintf(format, v...))
}

func (logger *Logger) D(format string, v ...interface{}) {
	logger.Debug(format, v...)
}

func (logger *Logger) Info(format string, v ...interface{}) {
	logger.genMsg(LevelInfo, fmt.Sprintf(format, v...))
}

func (logger *Logger) I(format string, v ...interface{}) {
	logger.Info(format, v...)
}

func (logger *Logger) Warn(format string, v ...interface{}) {
	logger.genMsg(LevelWarn, fmt.Sprintf(format, v...))
}

func (logger *Logger) Error(format string, v ...interface{}) {
	logger.genMsg(LevelError, fmt.Sprintf(format, v...))
}

func (logger *Logger) E(format string, v ...interface{}) {
	logger.Error(format, v...)
}

func (logger *Logger) Fatal(format string, v ...interface{}) {
	logger.genMsg(LevelFatal, fmt.Sprintf(format, v...))
}

func (logger *Logger) genMsg(level Level, log string) {
	if level < logger.lowestLevel {
		return
	}
	if atomic.LoadInt32(&logger.isRunning) == 0 {
		return
	}
	buf := logger.bufPool.Get()
	defer logger.bufPool.Release(buf)

	buf.Write(levelBytes[level])
	buf.WriteString(" ")
	buf.WriteString(time.Now().Format("2006-01-02 15:04:05.000"))
	buf.WriteString(" ")
	buf.WriteString(location(logger.callDepth + 3))
	buf.WriteString(" ")
	buf.WriteString(log)
	buf.WriteString("\n")

	select {
	case logger.msg <- &message{buf.String(), level}:
	default:
	}
}

func (logger *Logger) cleanMsg() {
	for {
		select {
		case msg := <-logger.msg:
			logger.Write(msg)
		default:
			return
		}
	}
}

func location(deep int) string {
	_, file, line, ok := runtime.Caller(deep)
	if !ok {
		file = "???"
		line = 0
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}
