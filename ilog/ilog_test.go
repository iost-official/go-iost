package ilog

import (
	"testing"
	"time"
)

func TestConsoleLog(t *testing.T) {
	Debug("this is a debug log")
	Info("this is a info log")
	Warn("this is a waining log")
	Error("this is a error log")
	Fatal("this is a fatal log")
	time.Sleep(time.Second)
}
