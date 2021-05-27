package ilog

import (
	"fmt"
	"os"
)

var _ LogWriter = &ConsoleWriter{}

func color(s string, level Level) string {
	return fmt.Sprintf("\033[1;%dm%s\033[0m", levelColor[level], s)
}

// ConsoleWriter is responsible for writing log to console.
type ConsoleWriter struct {
	level    Level
	colorful bool
}

// NewConsoleWriter returns a new instance of ConsoleWriter.
func NewConsoleWriter() *ConsoleWriter {
	colorful := false
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		colorful = true
	}
	return &ConsoleWriter{
		colorful: colorful,
	}
}

// Init inits ConsoleWriter.
func (cw *ConsoleWriter) Init() error {
	return nil
}

// SetLevel sets the log level.
func (cw *ConsoleWriter) SetLevel(l Level) {
	cw.level = l
}

// GetLevel returns the log level.
func (cw *ConsoleWriter) GetLevel() Level {
	return cw.level
}

// Write writes message to the console.
func (cw *ConsoleWriter) Write(msg string, level Level) error {
	if cw.colorful {
		msg = color(msg, level)
	}
	_, err := fmt.Fprint(os.Stdout, msg)
	return err
}

// Flush flushes.
func (cw *ConsoleWriter) Flush() error {
	return nil
}

// Close closes the writer.
func (cw *ConsoleWriter) Close() error {
	return nil
}
