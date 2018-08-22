package ilog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultLogger(t *testing.T) {
	Debug("this", "is a", "debug log")
	Info("this", "is a", "info log")
	Warn("this", "is a", "waining log")
	Error("this", "is a", "error log")
	Debugf("this is a %s log", "debug")
	Infof("this is a %s log", "info")
	Warnf("this is a %s log", "warning")
	Errorf("this is a %s log", "error")
	Flush()
}

func TestFileLogger(t *testing.T) {
	logger := New()
	fw := NewFileWriter("logs1/")
	err := logger.AddWriter(fw)
	assert.Nil(t, err)
	InitLogger(logger)

	Debug("this is a debug log")
	Info("this is a info log")
	Warn("this is a waining log")
	Error("this is a error log")
	Flush()
}

func TestAddWriter(t *testing.T) {
	fw := NewFileWriter("logs2/")
	err := AddWriter(fw)
	assert.Nil(t, err)

	Debug("this is a debug log")
	Info("this is a info log")
	Warn("this is a waining log")
	Error("this is a error log")
	Flush()
}

func BenchmarkFileLogger(b *testing.B) {
	logger := New()
	fw := NewFileWriter("benchlogs/")
	logger.AddWriter(fw)
	InitLogger(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark: %d", i)
	}
	logger.Flush()
}
