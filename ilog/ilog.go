package ilog

var (
	defaultLogger *Logger
)

func init() {
	defaultLogger = NewConsoleLogger()
	defaultLogger.SetCallDepth(1)
	defaultLogger.Start()
}

func InitLogger(logger *Logger) {
	defaultLogger.Stop()
	defaultLogger = logger
	defaultLogger.Start()
}

func AddWriter(lw LogWriter) error {
	return defaultLogger.AddWriter(lw)
}

func SetCallDepth(d int) {
	defaultLogger.SetCallDepth(d)
}

func Stop() {
	defaultLogger.Stop()
}

func Flush() {
	defaultLogger.Flush()
}

func Debug(format string, v ...interface{}) {
	defaultLogger.Debug(format, v...)
}

func D(format string, v ...interface{}) {
	defaultLogger.D(format, v...)
}

func Info(format string, v ...interface{}) {
	defaultLogger.Info(format, v...)
}

func I(format string, v ...interface{}) {
	defaultLogger.I(format, v...)
}

func Warn(format string, v ...interface{}) {
	defaultLogger.Warn(format, v...)
}

func Error(format string, v ...interface{}) {
	defaultLogger.Error(format, v...)
}

func E(format string, v ...interface{}) {
	defaultLogger.E(format, v...)
}

func Fatal(format string, v ...interface{}) {
	defaultLogger.Fatal(format, v...)
}
