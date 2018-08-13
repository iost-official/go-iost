package ilog

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

var _ LogWriter = &ConsoleWriter{}

func color(s string, level Level) string {
	return fmt.Sprintf("\033[1;%dm%s\033[0m", levelColor[level], s)
}

type ConsoleWriter struct {
	level    Level
	colorful bool
}

func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{
		colorful: terminal.IsTerminal(int(os.Stdout.Fd())),
	}
}

func (cw *ConsoleWriter) Init() error {
	return nil
}

func (cw *ConsoleWriter) SetLevel(l Level) {
	cw.level = l
}

func (cw *ConsoleWriter) GetLevel() Level {
	return cw.level
}

func (cw *ConsoleWriter) Write(msg string, level Level) error {
	if cw.colorful {
		msg = color(msg, level)
	}
	_, err := fmt.Fprint(os.Stdout, msg)
	return err
}

func (cw *ConsoleWriter) Flush() error {
	return nil
}

func (cw *ConsoleWriter) Close() error {
	return nil
}
